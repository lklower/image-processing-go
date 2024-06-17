package imagetor

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"sync"
)

// Number of color channels (RGBA).
const channels int = 4
const numWorkers int = 4

// ImageToTensor converts an image.Image to a 3D tensor of float64 values.
//
// The image is converted to a tensor with each element representing the normalized
// RGB and alpha values of the corresponding pixel.
//
// Args:
//
//	img: The image to convert.
//
// Returns:
//
//	A 3D tensor representing the image, where each element is a float64 value
//	representing the normalized RGB and alpha values of the corresponding pixel.
func ImageToTensor(img image.Image) [][][]float64 {
	var bounds image.Rectangle = img.Bounds()
	var width int = bounds.Max.X - bounds.Min.X
	var height int = bounds.Max.Y - bounds.Min.Y
	var tileWidth int = width / numWorkers

	tensor := make([][][]float64, height)
	for y := 0; y < height; y++ {
		tensor[y] = make([][]float64, width)
		for x := 0; x < width; x++ {
			tensor[y][x] = make([]float64, channels)
		}
	}

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(start, end int) {
			defer wg.Done()
			for y := 0; y < height; y++ {
				for x := start; x < end; x++ {
					r, g, b, a := img.At(x, y).RGBA()
					tensor[y][x][0] = float64(r) / 65535.0
					tensor[y][x][1] = float64(g) / 65535.0
					tensor[y][x][2] = float64(b) / 65535.0
					tensor[y][x][3] = float64(a) / 65535.0
				}

			}
		}(i*tileWidth, (i+1)*tileWidth)
	}

	wg.Wait() // Wait for all goroutines finish
	return tensor
}

// TensorToImage converts a 3D tensor of float64 values to an image.Image.
//
// The tensor is converted to an image with each element representing the
// RGB and alpha values of the corresponding pixel.
//
// Args:
//
//	tensor: The tensor to convert.
//
// Returns:
//
//	An image.Image representing the tensor, where each pixel's RGB and alpha
//	values are derived from the corresponding element in the tensor.
func TensorToImage(tensor [][][]float64) image.Image {
	height, width := len(tensor), len(tensor[0])
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var tileWidth int = width / numWorkers

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(start, end int) {
			defer wg.Done()
			for y := 0; y < height; y++ {
				for x := start; x < end; x++ {
					r := uint16(tensor[y][x][0] * 65535.0)
					g := uint16(tensor[y][x][1] * 65535.0)
					b := uint16(tensor[y][x][2] * 65535.0)
					a := uint16(tensor[y][x][3] * 65535.0)
					img.Set(x, y, color.RGBA64{r, g, b, a})
				}
			}
		}(i*tileWidth, (i+1)*tileWidth)
	}

	wg.Wait() // Wait for all goroutines finish
	return img
}

// Resize resizes a tensor using bilinear interpolation.
//
// The tensor is resized to the specified width and height, preserving the
// aspect ratio of the original tensor.
//
// Args:
//
//	tensor: A pointer to the tensor to resize.
//	width: The desired width of the resized tensor.
//	height: The desired height of the resized tensor.
//
func Resize(tensor *[][][]float64, width int, height int) {
	oldHeight, oldWidth := len(*tensor), len((*tensor)[0])
	tileHeight := height / numWorkers

	newTensor := make([][][]float64, height)
	for y := 0; y < height; y++ {
		newTensor[y] = make([][]float64, width)
		for x := 0; x < width; x++ {
			newTensor[y][x] = make([]float64, channels)
		}
	}

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(start, end int) {
			defer wg.Done()

			for y := start; y < end; y++ {
				for x := 0; x < width; x++ {
					oldX := float64(x) * float64(oldWidth) / float64(width)
					oldY := float64(y) * float64(oldHeight) / float64(height)

					x0 := int(oldX)
					y0 := int(oldY)
					dx := oldX - float64(x0)
					dy := oldY - float64(y0)

					// Skip if the surrounding pixels are out of bounds
					if x0 < 0 || x0 >= oldWidth-1 || y0 < 0 || y0 >= oldHeight-1 {
						continue
					}

					// Perform bilinear interpolation for each channel
					for c := 0; c < channels; c++ {
						newTensor[y][x][c] = (1-dx)*(1-dy)*(*tensor)[y0][x0][c] + dx*(1-dy)*(*tensor)[y0][x0+1][c] + (1-dx)*dy*(*tensor)[y0+1][x0][c] + dx*dy*(*tensor)[y0+1][x0+1][c]
					}
				}
			}
		}(i*tileHeight, (i+1)*tileHeight)
	}

	wg.Wait() // Wait for all goroutines finish
	*tensor = newTensor
}

// scaleFactor calculates the scaling factor for an overlay image to fit within a target image while maintaining aspect ratio.
//
// It determines the maximum scaling factor that allows the overlay to fit within the target image without exceeding its dimensions.
//
// Args:
//
//	target: The target image represented as a 3D tensor of float64.
//	overlay: The overlay image represented as a 3D tensor of float64.
//
// Returns:
//
//	The scaling factor as a float64.
func scaleFactor(target [][][]float64, overlay [][][]float64) float64 {
	var factor float64 = 1.0
	baseWidth, baseHeight := len(target[0]), len(target)
	overlayWidth, overlayHeight := len(overlay[0]), len(overlay)

	if overlayWidth > baseWidth || overlayHeight > baseHeight {
		scaleX := float64(baseWidth) / float64(overlayWidth)
		scaleY := float64(baseHeight) / float64(overlayHeight)
		factor = min(scaleX, scaleY)
	}

	return factor
}

// AddOverlay adds an overlay image to a target image with alpha blending.
//
// The overlay image is resized to fit within the target image, maintaining its aspect ratio.
// The overlay is then centered within the target image.
//
// Alpha blending is applied to the overlay, allowing the target image to show through.
//
// Args:
//
//	target: The target image represented as a 3D tensor of float64.
//	overlay: The overlay image represented as a 3D tensor of float64.
//
// Returns:
//
//	A new 3D tensor representing the combined image, or an error if the target or overlay is empty.
func AddOverlay(target [][][]float64, overlay *[][][]float64) ([][][]float64, error) {
	if len(target) == 0 || len(*overlay) == 0 {
		return nil, fmt.Errorf("target or overlay is empty")
	}

	targetWidth := len(target[0])
	targetHeight := len(target)

	factor := scaleFactor(target, *overlay)

	// Calculate new overlay size
	newOverlayWidth := int(float64(len((*overlay)[0])) * factor)
	newOverlayHeight := int(float64(len(*overlay)) * factor)

	Resize(overlay, newOverlayWidth, newOverlayHeight)

	// Calculate center position for overlay
	offsetX := (targetWidth - newOverlayWidth) / 2
	offsetY := (targetHeight - newOverlayHeight) / 2

	// Create new 3d tensor to hold the result
	newTensor := make([][][]float64, targetHeight)
	for y := 0; y < targetHeight; y++ {
		newTensor[y] = make([][]float64, targetWidth)
		for x := 0; x < targetWidth; x++ {
			newTensor[y][x] = make([]float64, 4)
			copy(newTensor[y][x], target[y][x])
		}
	}

	tileHeight := newOverlayHeight / numWorkers
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		// startY := offsetY + i*tileHeight
		// endY := offsetY + (i+1)*tileHeight
		// if i == numWorkers-1 {
		// 	endY = offsetY + newOverlayHeight
		// }
		go func(startY, endY int) {
			defer wg.Done()
			// Apply overlay with alpha blending
			for y := startY; y < endY; y++ {
				for x := offsetX; x < offsetX+newOverlayWidth; x++ {
					alpha := (*overlay)[y-offsetY][x-offsetX][3]
					for i := 0; i < 3; i++ {
						newTensor[y][x][i] = (*overlay)[y-offsetY][x-offsetX][i] + ((1 - alpha) * newTensor[y][x][i])
					}
					newTensor[y][x][3] = 1.0
				}
			}
		}(offsetY+i*tileHeight, offsetY+(i+1)*tileHeight)
	}

	wg.Wait()
	return newTensor, nil
}

// UpSideDown flips the image represented by the tensor vertically.
//
// The function modifies the input tensor in place, flipping the image vertically.
//
// Args:
//
//	tensor: A pointer to the 3D tensor representing the image.
func UpSideDown(tensor *[][][]float64) {
	height, width := len(*tensor), len((*tensor)[0])

	for y := 0; y < height/2; y++ {
		tr := (*tensor)[y]
		br := (*tensor)[height-1-y]

		for x := 0; x < width; x++ {
			tr[x], br[x] = br[x], tr[x]
		}
	}
}

// GrayScale converts the image represented by the tensor to grayscale.
//
// The function modifies the input tensor in place, converting the image to grayscale
// using the LUMINOSITY method.
//
// Args:
//
//	tensor: A pointer to the 3D tensor representing the image.
func GrayScale(tensor *[][][]float64) {
	height, width := len(*tensor), len((*tensor)[0])

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b := (*tensor)[y][x][0], (*tensor)[y][x][1], (*tensor)[y][x][2]

			// Calculate grayscale value using LUMINOSITY method
			gray := 0.2126*r + 0.7152*g + 0.0722*b

			(*tensor)[y][x][0] = gray
			(*tensor)[y][x][1] = gray
			(*tensor)[y][x][2] = gray
		}
	}

}

// Rotate rotates the image represented by the tensor by the specified angle.
//
// The function modifies the input tensor in place, rotating the image by the
// specified angle using bilinear interpolation.
//
// Args:
//
//	tensor: A pointer to the 3D tensor representing the image.
//	angle: The angle to rotate the image by, in degrees.
func Rotate(tensor *[][][]float64, angle float64) {
	height, width := len(*tensor), len((*tensor)[0])

	// Calculate Center
	centerX, centerY := float64(width)/2.0, float64(height)/2.0

	//Convert Angle to Radians
	radians := angle * math.Pi / 180

	tempTensor := make([][][]float64, height)
	for y := 0; y < height; y++ {
		tempTensor[y] = make([][]float64, width)
		for x := 0; x < width; x++ {
			tempTensor[y][x] = make([]float64, channels)
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotateX := float64(x) - centerX
			rotateY := float64(y) - centerY

			originalX := rotateX*math.Cos(radians) + rotateY*math.Sin(radians) + centerX
			originalY := -rotateX*math.Sin(radians) + rotateY*math.Cos(radians) + centerY

			// Bilinear Interpolation (if within bounds)
			if originalX >= 0 && originalX < float64(width) && originalY >= 0 && originalY < float64(height) {
				x1, y1 := int(math.Floor(originalX)), int(math.Floor(originalY))
				x2, y2 := x1, y1
				dx, dy := originalX-math.Floor(originalX), originalY-math.Floor(originalY)

				for c := 0; c < channels; c++ {
					tempTensor[y][x][c] = (1-dx)*(1-dy)*(*tensor)[y1][x1][c] +
						dx*(1-dy)*(*tensor)[y1][x2][c] +
						(1-dx)*dy*(*tensor)[y2][x1][c] +
						dx*dy*(*tensor)[y2][x2][c]
				}
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			(*tensor)[y][x] = tempTensor[y][x]
		}
	}
}

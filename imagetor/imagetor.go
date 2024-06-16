package imagetor

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"sync"
)

// Number of color channels (RGBA).
const channels int = 4
const numWorkers int = 4

// imageToTensor converts an image.Image to a 3D tensor of float64.
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

// tensorToImage converts a 3D tensor of float64 to an image.Image.
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

// resize resizes a tensor using bilinear interpolation.
func Resize(tensor [][][]float64, width int, height int) [][][]float64 {
	oldHeight, oldWidth := len(tensor), len(tensor[0])
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
						newTensor[y][x][c] = (1-dx)*(1-dy)*tensor[y0][x0][c] + dx*(1-dy)*tensor[y0][x0+1][c] + (1-dx)*dy*tensor[y0+1][x0][c] + dx*dy*tensor[y0+1][x0+1][c]
					}
				}
			}
		}(i*tileHeight, (i+1)*tileHeight)
	}

	wg.Wait() // Wait for all goroutines finish
	return newTensor
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
func AddOverlay(target [][][]float64, overlay [][][]float64) ([][][]float64, error) {
	if len(target) == 0 || len(overlay) == 0 {
		return nil, fmt.Errorf("target or overlay is empty")
	}

	targetWidth := len(target[0])
	targetHeight := len(target)

	factor := scaleFactor(target, overlay)

	// Calculate new overlay size
	newOverlayWidth := int(float64(len(overlay[0])) * factor)
	newOverlayHeight := int(float64(len(overlay)) * factor)

	scaleOverlay := Resize(overlay, newOverlayWidth, newOverlayHeight)

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
					alpha := scaleOverlay[y-offsetY][x-offsetX][3]
					for i := 0; i < 3; i++ {
						newTensor[y][x][i] = scaleOverlay[y-offsetY][x-offsetX][i] + ((1 - alpha) * newTensor[y][x][i])
					}
					newTensor[y][x][3] = 1.0
				}
			}
		}(offsetY+i*tileHeight, offsetY+(i+1)*tileHeight)
	}

	wg.Wait()
	return newTensor, nil
}

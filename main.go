package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"os"
	"sync"
)

// Imagetor struct to hold methods for image manipulation.
type Imagetor struct{}

// Number of color channels (RGBA).
const channels int = 4

// imageToTensor converts an image.Image to a 3D tensor of float64.
func (i *Imagetor) imageToTensor(img image.Image) [][][]float64 {
	var bounds image.Rectangle = img.Bounds()
	var width int = bounds.Max.X - bounds.Min.X
	var height int = bounds.Max.Y - bounds.Min.Y
	tensor := make([][][]float64, height)

	for y := 0; y < height; y++ {
		tensor[y] = make([][]float64, width)
		for x := 0; x < width; x++ {
			tensor[y][x] = make([]float64, channels)
			r, g, b, a := img.At(x, y).RGBA()
			tensor[y][x][0] = float64(r) / 65535.0
			tensor[y][x][1] = float64(g) / 65535.0
			tensor[y][x][2] = float64(b) / 65535.0
			tensor[y][x][3] = float64(a) / 65535.0
		}
	}
	return tensor
}

// tensorToImage converts a 3D tensor of float64 to an image.Image.
func (i *Imagetor) tensorToImage(tensor [][][]float64) image.Image {
	height, width := len(tensor), len(tensor[0])
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint16(tensor[y][x][0] * 65535.0)
			g := uint16(tensor[y][x][1] * 65535.0)
			b := uint16(tensor[y][x][2] * 65535.0)
			a := uint16(tensor[y][x][3] * 65535.0)
			img.Set(x, y, color.RGBA64{r, g, b, a})
		}
	}
	return img
}

// resize resizes a tensor using bilinear interpolation.
func (i *Imagetor) resize(tensor [][][]float64, width int, height int) [][][]float64 {
	oldHeight, oldWidth := len(tensor), len(tensor[0])

	newTensor := make([][][]float64, height)
	for y := 0; y < height; y++ {
		newTensor[y] = make([][]float64, width)
		for x := 0; x < width; x++ {
			newTensor[y][x] = make([]float64, channels)
		}
	}

	var wg sync.WaitGroup
	wg.Add(height)

	// Process each row concurrently
	for y := 0; y < height; y++ {
		go func(y int) {
			defer wg.Done()
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
		}(y)
	}

	wg.Wait() // Wait for all goroutines finish
	return newTensor
}

func main() {
	f, e := os.Open("image.jpg")
	if e != nil {
		fmt.Println("Failed to open image: ", e)
		return
	}
	defer f.Close()

	img, _, e := image.Decode(f)
	if e != nil {
		fmt.Println("Failed to decode image: ", e)
		return
	}

	imagetor := Imagetor{}

	tensor := imagetor.imageToTensor(img)

	tensor = imagetor.resize(tensor, 460, 260)

	newImg := imagetor.tensorToImage(tensor)

	file, err := os.Create("new.jpg")
	if err != nil {
		fmt.Println("Error creating new image: ", err)
	}
	defer file.Close()

	if e := jpeg.Encode(file, newImg, &jpeg.Options{Quality: 100}); e != nil {
		fmt.Println("Failed to encode image: ", e)
		return
	}

	fmt.Println("created new iamge successfully!")
}

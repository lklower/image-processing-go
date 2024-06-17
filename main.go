package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"mymodule/imagetor"
	"os"
	"time"
)

func isJPGorPNG(file *os.File) bool {
	var header [8]byte
	if _, err := file.Read(header[:]); err != nil {
		panic(err)
	}
	file.Seek(0, 0)

	switch {
	case bytes.Equal(header[:2], []byte{0xFF, 0xD8}): // JPEG
		return true
	case bytes.Equal(header[:], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}): // PNG
		return true
	default:
		return false
	}
}
func openImage(path string) (image.Image, error) {
	file, e := os.Open(path)
	if e != nil {
		fmt.Println("Failed to open image: ", e)
		return nil, e
	}
	defer file.Close()

	if !isJPGorPNG(file) {
		fmt.Println("Unsupported image format")
		return nil, fmt.Errorf("unsupported image format")
	}

	img, _, e := image.Decode(file)
	if e != nil {
		fmt.Println("Failed to decode image: ", e)
		return nil, e
	}
	return img, nil
}

func saveImage(result image.Image, path string) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	if err := jpeg.Encode(writer, result, &jpeg.Options{Quality: 100}); err != nil {
		return err
	}
	return nil
}

func main() {

	startTime := time.Now()

	targetImage, err := openImage("large-image.jpg")
	if err != nil {
		fmt.Println("Error opening image: ", err)
		return
	}

	logoImage, err := openImage("logo.png")
	if err != nil {
		fmt.Println("Error opening image: ", err)
		return
	}

	targetTensor := imagetor.ImageToTensor(targetImage)
	logoTensor := imagetor.ImageToTensor(logoImage)

	if err := imagetor.AddOverlay(&targetTensor, &logoTensor); err != nil {
		fmt.Println("Error adding overlay: ", err)
		return
	}

	//imagetor.UpSideDown(&resultTensor)

	// imagetor.GrayScale(&resultTensor)

	// imagetor.Rotate(&resultTensor, 5.0)

	resultImage := imagetor.TensorToImage(targetTensor)

	_ = saveImage(resultImage, "output.jpg")

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	fmt.Println("Image saved successfully.")

	fmt.Println("Elapsed time: ", elapsedTime)
}

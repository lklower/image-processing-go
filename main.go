package main

import (
	"bufio"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"mymodule/imagetor"
	"os"
	"time"
)

func openImage(path string) (image.Image, error) {
	file, e := os.Open(path)
	if e != nil {
		fmt.Println("Failed to open image: ", e)
		return nil, e
	}
	defer file.Close()

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

	resultTensor, err := imagetor.AddOverlay(targetTensor, logoTensor)
	if err != nil {
		fmt.Println("Error adding overlay: ", err)
		return
	}

	resultImage := imagetor.TensorToImage(resultTensor)

	saveImage(resultImage, "output.jpg")

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	fmt.Println("Image saved successfully.")

	fmt.Println("Elapsed time: ", elapsedTime)
}

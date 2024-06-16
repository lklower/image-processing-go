package main

import (
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

	writer, err := os.Create("output.jpg")
	if err != nil {
		fmt.Println("Error creating output file: ", err)
		return
	}
	defer writer.Close()

	if err := jpeg.Encode(writer, resultImage, &jpeg.Options{Quality: 100}); err != nil {
		fmt.Println("Error encoding image: ", err)
		return
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	fmt.Println("Image saved successfully.")

	fmt.Println("Elapsed time: ", elapsedTime)
}

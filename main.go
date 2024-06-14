package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"mymodule/imagetor"
	"os"
)

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

	tensor := imagetor.ImageToTensor(img)

	tensor = imagetor.Resize(tensor, 1920, 1080)

	newImg := imagetor.TensorToImage(tensor)

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

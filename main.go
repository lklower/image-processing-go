package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"mymodule/imagetor"
	"os"
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
	img, err := openImage("image.jpg")
	if err != nil {
		fmt.Println("Error opening image: ", err)
		return
	}

	tensor := imagetor.ImageToTensor(img)

	// tensor = imagetor.Resize(tensor, 1920, 1080)

	// newImg := imagetor.TensorToImage(tensor)

	// file, err := os.Create("new.jpg")
	// if err != nil {
	// 	fmt.Println("Error creating new image: ", err)
	// }
	// defer file.Close()

	// if e := jpeg.Encode(file, newImg, &jpeg.Options{Quality: 100}); e != nil {
	// 	fmt.Println("Failed to encode image: ", e)
	// 	return
	// }

	// fmt.Println("created new iamge successfully!")

	logo, err := openImage("logo.png")
	if err != nil {
		fmt.Println("Error opening image: ", err)
		return
	}

	logoTensor := imagetor.ImageToTensor(logo)

	// factor := scaleFactor(tensor, logoTensor)
	// newLogoWidth := int(float64(len(logoTensor[0])) * factor)
	// newLogoHeight := int(float64(len(logoTensor)) * factor)

	// logoTensor = imagetor.Resize(logoTensor, newLogoWidth, newLogoHeight)

	overlayTensor, err := imagetor.AddOverlay(tensor, logoTensor)
	if err != nil {
		fmt.Println("Error adding overlay: ", err)
		return
	}

	result := imagetor.TensorToImage(overlayTensor)

	writer, err := os.Create("final_result.jpg")
	if err != nil {
		fmt.Println("Error creating new image: ", err)
	}
	defer writer.Close()

	if e := jpeg.Encode(writer, result, &jpeg.Options{Quality: 100}); e != nil {
		fmt.Println("Failed to encode image: ", e)
		return
	}

	fmt.Println("created new logo successfully!")
}

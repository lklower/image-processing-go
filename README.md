# Image Watermarking with Go and imagetor

This project demonstrates a simple image watermarking application using Go and the imagetor module. It takes two images as input: a target image and a watermark image (logo). The application then overlays the watermark onto the target image and saves the resulting image.

# Features:

Image Loading and Saving: The project utilizes Go's built-in image library to load and save images in various formats (JPEG, PNG).
Image Manipulation: The imagetor module provides functions to convert images to tensors and perform overlay operations.
Performance Measurement: The code tracks the execution time to provide insights into the performance of the watermarking process.

# Dependencies:

imagetor: This module is assumed to be a custom module providing image manipulation functions. You will need to install and configure it according to its documentation.
Usage:

Install Go: Ensure you have Go installed on your system.
Install Dependencies: Install the imagetor module (or any other required dependencies).
Prepare Images: Place the target image (`large-image.jpg`) and the watermark image (`logo.png`) in the same directory as the main.go file.
Run the Application: Execute the main.go file using the go run command.
Output: The resulting watermarked image will be saved as output.jpg in the same directory.

# Code Breakdown:

openImage function: Loads an image from a given path and returns an image.Image object.
saveImage function: Saves an image.Image object to a specified path in JPEG format.
main function:
Loads the target image and the watermark image.
Converts both images to tensors using `imagetor.ImageToTensor`.
Overlays the watermark tensor onto the target tensor using `imagetor.AddOverlay`.
Converts the resulting tensor back to an image using `imagetor.TensorToImage`.
Saves the watermarked image to `output.jpg`.
Measures and prints the execution time.

# Customization:

Watermark Image: You can replace `logo.png` with any desired watermark image.
Output Format: Modify the saveImage function to save the output in a different format (e.g., PNG).
Overlay Position: The imagetor module likely provides options to adjust the position and size of the watermark.

# Note:

This project serves as a basic example of image watermarking in Go using the imagetor module. It can be extended to include more advanced features like transparency control, watermark resizing, and multiple watermarking.

**This project provides a starting point for building a more comprehensive image watermarking application using Go and the imagetor module.**


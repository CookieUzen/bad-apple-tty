package main

import (
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

func main() {

	// load a folder
	images, err := loadFolder("images_small")
	if err != nil {
		panic(err)
	}

	fps := 30
	interval := time.Second / time.Duration(fps)

	// loop through every image in the folder
	for _, file := range images {
		start := time.Now()

		// Clear the terminal
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		// Open the image
		image, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		// Import the image
		pixels, height, width, err := importFrame(image)
		if err != nil {
			panic(err)
		}

		// Quantize the image
		pixels = quantize(pixels, height, width, 128)

		// Print the image to the terminal using full blocks
		printFullBlocks(pixels, height, width)
		
		// Sleep for the remainder of the frame
		elapsed := time.Since(start)

		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}
	}

}

// loads all png in a folder into an array of images
func loadFolder(folder string) ([]string, error) {
	
	// get all the files in the folder
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	// create an array to hold the images
	images := make([]string, len(files))

	// loop through the files and add them to the array
	for i, file := range files {
		images[i] = folder + "/" + file.Name()
	}

	return images, nil
}



// Reads an image and returns a 2D array of the pixels in grayscale
func importFrame(image io.Reader) ([][]uint8, int, int, error) {

	// Open the image
	file, err := png.Decode(image)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get the image bounds
	bounds := file.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Create a 2D array of the pixels
	pixels := make([][]uint8, height)
	for i := range pixels {
		pixels[i] = make([]uint8, width)
	}

	// Iterate over the pixels and convert them to grayscale
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels[y][x] = color.GrayModel.Convert(file.At(x, y)).(color.Gray).Y
		}
	}

	return pixels, height, width, nil
}

// convert an grayscale image to a quantized array
func quantize(pixels [][]uint8, height int, width int, threshold uint8) [][]uint8 {

	output := make([][]uint8, height)
	for i := range output {
		output[i] = make([]uint8, width)
	}
	
	// loop through the pixels and quantize them
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if pixels[y][x] < threshold {
				output[y][x] = 0
			} else {
				output[y][x] = 1
			}
		}
	}

	return output
}

// Print out the image using full blocks, doubling width for perfect pixels
func printFullBlocks(pixels [][]uint8, height int, width int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if pixels[y][x] == 0 {
				fmt.Print("██")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}


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

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	// load a folder
	images, err := loadFolder("images")
	if err != nil {
		panic(err)
	}

	fps := 30
	interval := time.Second / time.Duration(fps)

	// create a buffer for the first image to get the size
	image, err := os.Open(images[0])
	if err != nil {
		panic(err)
	}

	// Import the images
	lastImage, height, width, err := importFrame(image)
	if err != nil {
		panic(err)
	}

	// Quantize the images
	lastImage = quantize(lastImage, height, width, 170)

	// Print the images to the terminal using half printFullBlocks
	printHalfBlocks(lastImage, height, width)

	// loop through the rest of the image in the folder
	for i, file := range images {
		// Skip the first image
		if i == 0 {
			continue
		}

		// Start the timer
		start := time.Now()

		// Reset the cursor
		fmt.Println("\033[1;1H")

		fmt.Println("clear: ", time.Since(start))

		// Open the image
		image, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		fmt.Println("Open: ", time.Since(start))

		// Import the image
		pixels, height, width, err := importFrame(image)
		if err != nil {
			panic(err)
		}

		fmt.Println("Import: ", time.Since(start))

		// Quantize the image
		pixels = quantize(pixels, height, width, 170)

		fmt.Println("Quantize: ", time.Since(start))

		// Print the image difference to the terminal
		printHalfBlocksDiff(pixels, height, width, lastImage)

		fmt.Println("Print: ", time.Since(start))

		lastImage = pixels
		
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

// subsample an image by a factor of 2 vertically
func subsample(pixels [][]uint8, height int, width int) ([][]uint8, int, int) {

	// int truncate will make sure array is even
	height /= 2
	
	output := make([][]uint8, height)
	for i := range output {
		output[i] = make([]uint8, width)
	}

	// loop through the pixels and subsample them
	for y := 0; y < height; y ++ {
		for x := 0; x < width; x++ {
			output[y][x] = pixels[y*2][x]/2 + pixels[y*2+1][x]/2
		}
	}

	return output, height, width
}


// Print out the image using full blocks, doubling width for perfect pixels
func printFullBlocks(pixels [][]uint8, height int, width int, repeat int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			for i := 0; i < repeat; i++ {
				if pixels[y][x] == 0 {
					fmt.Print("█")
				} else {
					fmt.Print("\033[1C")
				}
			}
		}
		fmt.Println()
	}
}

// Print out the image using half blocks
func printHalfBlocks(pixels [][]uint8, height int, width int) {
	for y := 0; y < height/2; y++ {
		for x := 0; x < width; x++ {
			// if both pixels are black
			if pixels[y*2][x] == 0 && pixels[y*2+1][x] == 0 {
				fmt.Print("█")
				continue
			}

			// if both pixels are white
			if pixels[y*2][x] == 1 && pixels[y*2+1][x] == 1 {
				fmt.Print(" ")
				continue
			}

			// if the top pixel is black and the bottom is white
			if pixels[y*2][x] == 0 && pixels[y*2+1][x] == 1 {
				fmt.Print("▀")
				continue
			}

			// if the top pixel is white and the bottom is black
			if pixels[y*2][x] == 1 && pixels[y*2+1][x] == 0 {
				fmt.Print("▄")
				continue
			}
		}
		fmt.Println()
	}
}

// Print out only the pixels that changed using half blocks
func printHalfBlocksDiff(pixels [][]uint8, height int, width int, lastPixels [][]uint8) {
	for y := 0; y < height/2; y++ {
		for x := 0; x < width; x++ {
			// if both pixels are black
			if pixels[y*2][x] == 0 && pixels[y*2+1][x] == 0 {
				if lastPixels[y*2][x] == 0 && lastPixels[y*2+1][x] == 0 {
					fmt.Print("\033[1C")
				} else {
					fmt.Print("█")
				}
				continue
			}

			// if both pixels are white
			if pixels[y*2][x] == 1 && pixels[y*2+1][x] == 1 {
				if lastPixels[y*2][x] == 1 && lastPixels[y*2+1][x] == 1 {
					fmt.Print("\033[1C")
				} else {
					fmt.Print(" ")
				}
				continue
			}

			// if the top pixel is black and the bottom is white
			if pixels[y*2][x] == 0 && pixels[y*2+1][x] == 1 {
				if lastPixels[y*2][x] == 0 && lastPixels[y*2+1][x] == 1 {
					fmt.Print("\033[1C")
				} else {
					fmt.Print("▀")
				}
				continue
			}

			// if the top pixel is white and the bottom is black
			if pixels[y*2][x] == 1 && pixels[y*2+1][x] == 0 {
				if lastPixels[y*2][x] == 1 && lastPixels[y*2+1][x] == 0 {
					fmt.Print("\033[1C")
				} else {
					fmt.Print("▄")
				}
				continue
			}
		}
		fmt.Println()
	}
}

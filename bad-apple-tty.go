package main

import (
	"fmt"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
	"unicode/utf8"
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

	// loop through every image in the folder
	for _, file := range images {
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
		pixels = quantize(pixels, height, width, 128)

		fmt.Println("Quantize: ", time.Since(start))

		// Print the image to the terminal using half blocks
		printHalfBlocks(pixels, height, width)

		fmt.Println("Print: ", time.Since(start))
		
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
	// Use a buffer of bytes to store the printed output
	buffer := make([]byte, 0, width*height/2*(utf8.UTFMax+1))

	// Define the byte representation of the different block types and characters
	space := []byte(" ")
	topBlock := []byte("▀")
	bottomBlock := []byte("▄")
	fullBlock := []byte("█")
	newline := []byte("\n")

	// Loop through each row of pixels
	for y := 0; y < height/2; y++ {
		// Loop through each column of pixels in the current row
		for x := 0; x < width; x++ {
			// Determine the block type to print based on the values of the two pixels
			var blockType []byte
			if pixels[y*2][x] == 0 && pixels[y*2+1][x] == 0 {
				blockType = fullBlock
			} else if pixels[y*2][x] == 1 && pixels[y*2+1][x] == 1 {
				blockType = space
			} else if pixels[y*2][x] == 0 && pixels[y*2+1][x] == 1 {
				blockType = topBlock
			} else if pixels[y*2][x] == 1 && pixels[y*2+1][x] == 0 {
				blockType = bottomBlock
			}
			// Append the block type to the buffer
			buffer = append(buffer, blockType...)
		}
		// Append a newline character after each row of blocks
		buffer = append(buffer, newline...)
	}

	// Print the contents of the buffer as a string
	fmt.Print(string(buffer))
}


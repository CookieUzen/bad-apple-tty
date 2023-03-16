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
	"strconv"
	"flag"
	"strings"
)

// For parsing command line arguments
var fps int
var mode string
var args []string
var threshold int

func init() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [options] [arguments]\n", os.Args[0])
		fmt.Println("-args: \nFolder for images to load")
        flag.PrintDefaults()
    }

	// Parse the command line arguments
	flag.IntVar(&fps, "f", 30, "fps to run at")
	flag.StringVar(&mode, "m", "truecolor", "mode to run in (tty, unicode, truecolor")
	flag.IntVar(&threshold, "t", 128, "threshold for quantization")

	flag.Parse()

	args = flag.Args()
}

func main() {
	
	// Sanity check the arguments
	if len(args) < 1 {
		fmt.Println("Wrong number of arguments")
		flag.Usage()
		os.Exit(1)
	}

	folder := args[0]

	if fps < 1 {
		fmt.Println("Invalid fps")
		flag.Usage()
		os.Exit(1)
	}

	if threshold < 0 || threshold > 255 {
		fmt.Println("Invalid threshold")
		flag.Usage()
		os.Exit(1)
	}

	// Clear the screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	// load a folder
	images, err := loadFolder(folder)
	if err != nil {
		panic(err)
	}

	interval := time.Second / time.Duration(fps)

	// hide cursor
	fmt.Print("\033[?25l")

	// Parse the flags
	switch mode {
		case "tty":
			runTty(images, interval)
		case "tty-subsample":
			runTtySubsample(images, interval)
		case "unicode":
			runUnicode(images, interval)
		case "truecolor":
			runTruecolor(images, interval)
		default:
			fmt.Println("Invalid mode")
			flag.Usage()
			os.Exit(1)
	}


	// show cursor
	fmt.Print("\033[?25h")
}

// Run in tty mode
func runTty(images []string, interval time.Duration) {
	// loop through every image in the folder
	for _, file := range images {
		start := time.Now()

		// Reset the cursor
		fmt.Println("\033[H")

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
		quantized := quantize(pixels, height, width, uint8(threshold))

		// Print the image
		printFullBlocks(quantized, height, width, 1)
		
		// Sleep for the remainder of the frame
		elapsed := time.Since(start)

		// Print the frame rate
		fmt.Println("Theoretical FPS: " + strconv.Itoa(int(time.Second / elapsed)))

		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}

		elapsed = time.Since(start)
		fmt.Println("Real FPS: " + strconv.Itoa(int(time.Second / elapsed)))
	}

}

// Run in tty-subsample mode
func runTtySubsample(images []string, interval time.Duration) {
	// loop through every image in the folder
	for _, file := range images {
		start := time.Now()

		// Reset the cursor
		fmt.Println("\033[H")

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

		// Subsample the image
		subsampled, height, width := subsample(pixels, height, width)

		// Quantize the image
		quantized := quantize(subsampled, height, width, uint8(threshold))

		// Print the image
		printFullBlocks(quantized, height, width, 1)
		
		// Sleep for the remainder of the frame
		elapsed := time.Since(start)

		// Print the frame rate
		fmt.Println("Theoretical FPS: " + strconv.Itoa(int(time.Second / elapsed)))

		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}

		elapsed = time.Since(start)
		fmt.Println("Real FPS: " + strconv.Itoa(int(time.Second / elapsed)))
	}

}

// Run in unicode mode
func runUnicode(images []string, interval time.Duration) {
	// loop through every image in the folder
	for _, file := range images {
		start := time.Now()

		// Reset the cursor
		fmt.Println("\033[H")

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
		quantized := quantize(pixels, height, width, uint8(threshold))

		// Print the image
		printHalfBlocks(quantized, height, width)
		
		// Sleep for the remainder of the frame
		elapsed := time.Since(start)

		// Print the frame rate
		fmt.Println("Theoretical FPS: " + strconv.Itoa(int(time.Second / elapsed)))

		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}

		elapsed = time.Since(start)
		fmt.Println("Real FPS: " + strconv.Itoa(int(time.Second / elapsed)))
	}

}


// Run in truecolor mode
func runTruecolor(images []string, interval time.Duration) {
	// loop through every image in the folder
	for _, file := range images {
		start := time.Now()

		// Reset the cursor
		fmt.Println("\033[H")

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

		printHalfBlocksColor(pixels, height, width)
		
		// Sleep for the remainder of the frame
		elapsed := time.Since(start)

		// Print the frame rate
		fmt.Println("Theoretical FPS: " + strconv.Itoa(int(time.Second / elapsed)))

		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}

		elapsed = time.Since(start)
		fmt.Println("Real FPS: " + strconv.Itoa(int(time.Second / elapsed)))
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


// Print out the image using colored spaces, doubling width for perfect pixels
func printFullBlocks(pixels [][]uint8, height int, width int, repeat int) {
	// Use a buffer of bytes to store the printed output
	buffer := make([]byte, 0, width*height*8*2)
	black := []byte("\033[40m")
	white := []byte("\033[47m")
	reset := []byte("\033[0m")
	newline := []byte("\n")

	// repeat space to make a square (in case we don't subsample)
	space := []byte(strings.Repeat(" ", repeat))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if pixels[y][x] == 0 {
				buffer = append(buffer, black...)
			} else {
				buffer = append(buffer, white...)
			}
			buffer = append(buffer, space...)
		}

		buffer = append(buffer, reset...)
		buffer = append(buffer, newline...)
	}

	fmt.Print(string(buffer))
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

// Print half blocks with a color gradient
// done using true color escape sequences with grayscale values
func printHalfBlocksColor(pixels [][]uint8, height int, width int) {
	// Use a buffer of bytes to store the printed output
	buffer := make([]byte, 0, width*height/2*(utf8.UTFMax+1 + 20*8 + 100))

	// Define the byte representation of the different block types and characters
	block := []byte("▀")
	newline := []byte("\n")

	// Define color escape sequences
	reset := []byte("\033[0m")

	// Loop through each row of pixels
	for y := 0; y < height/2; y++ {

		// cache the first pixel in the row
		last_top := pixels[y*2][0]
		last_bottom := pixels[y*2+1][0]

		// Loop through each column of pixels in the current row
		for x := 0; x < width; x++ {

			// Determine the color to print based on the value of the pixel
			top := pixels[y*2][x]
			bottom := pixels[y*2+1][x]

			// foreground determines the top half of the block
			if x == 0 || last_top != top {

				top_char := []byte(strconv.Itoa(int(top)))

				// Save time by not using fmt.Sprintf
				// fmt.Sprintf("\033[38;2;%d;%d;%dm", top, top, top)
				buffer = append(buffer, []byte("\033[38;2;")...)
				buffer = append(buffer, top_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, top_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, top_char...)
				buffer = append(buffer, []byte("m")...)

				last_top = top
			}

			// background determines the bottom half of the block
			if x == 0 || last_bottom != bottom {

				bottom_char := []byte(strconv.Itoa(int(bottom)))

				// Save time by not using fmt.Sprintf
				// fmt.Sprintf("\033[48;2;%d;%d;%dm", bottom, bottom, bottom)
				buffer = append(buffer, []byte("\033[48;2;")...)
				buffer = append(buffer, bottom_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, bottom_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, bottom_char...)
				buffer = append(buffer, []byte("m")...)

				last_bottom = bottom
			}
			
			// Append the block type to the buffer
			buffer = append(buffer, block...)

		}
		// Append a newline character after each row of blocks
		buffer = append(buffer, newline...)

		// Append the reset escape sequence to the buffer
		buffer = append(buffer, reset...)
	}

	// Print the contents of the buffer as a string
	fmt.Print(string(buffer))
}

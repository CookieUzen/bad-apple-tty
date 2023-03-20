package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"
	"unicode/utf8"
	"strconv"
	"flag"
	"strings"
	"os/signal"
	"gocv.io/x/gocv"
	"golang.org/x/crypto/ssh/terminal"
)

// For parsing command line arguments
var fps int
var mode string
var args []string
var threshold int
var skipFrame bool
var terminalWidth int
var terminalHeight int

func init() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [options] [arguments]\n", os.Args[0])
		fmt.Println("-args: \nVideo to load")
        flag.PrintDefaults()
    }

	// Parse the command line arguments
	flag.IntVar(&fps, "f", 30, "fps to run at")
	flag.StringVar(&mode, "m", "truecolor", "mode to run in (tty, tty_subsample, unicode, truecolor)")
	flag.IntVar(&threshold, "t", 128, "threshold for quantization")
	flag.BoolVar(&skipFrame, "s", true, "skip frames if we're running too slow")

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

	// In case of a ^C, cleanup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		for sig := range c {
			// sig is a ^C, handle it
			if sig.String() == "interrupt" {
				cleanup()
				fmt.Println("Interrupted")
			} else {
				panic(sig)
			}
			
			os.Exit(1)
		}
	}()

	// Clear the screen
	fmt.Print("\033[2J")

	// Hide the cursor
	fmt.Print("\033[?25l")

	// Run the program
	videoMode(args[0])

	// Cleanup
	cleanup()

}

func cleanup() {
	// Show the cursor
	fmt.Print("\033[?25h")

	// Clear the screen
	fmt.Print("\033[2J")

	// Reset the cursor position
	fmt.Print("\033[0;0H")
}

// Runs the program by reading from a video
func videoMode (fileName string) {
	// Open the video
	video, err := gocv.VideoCaptureFile(fileName)
	if err != nil {
		fmt.Println("Error opening video")
		os.Exit(1)
	}
	defer video.Close()

	// Calculate the frame interval
	interval := time.Second / time.Duration(fps)

	// Start timer for the video
	videoTimer := time.Now()

	// Loop through the video
	for frameCount := uint64(0); ; frameCount++ {
		// Check if we're running too slow
		currentTime := int64(frameCount) * interval.Nanoseconds()
		if skipFrame && time.Since(videoTimer).Nanoseconds() > currentTime {
			// We're running too slow, skip frames
			framesToSkip := int(uint64(time.Since(videoTimer).Nanoseconds() / interval.Nanoseconds()) - frameCount)
			video.Grab(framesToSkip)

			// Update the frame count
			frameCount += uint64(framesToSkip)
		}

		// Start the timer
		start := time.Now()

		// Reset the cursor position using ansi
		fmt.Print("\033[0;0H")

		// Create a frame
		frame := gocv.NewMat()
		ok := video.Read(&frame)

		if !ok {
			// end of the video
			break
		}

		// Close the frame when done
		defer frame.Close()

		processImage(frame)

		// Print the frame rate
		elapsed := time.Since(start)
		fmt.Println("Theoretical FPS: " + strconv.Itoa(int(time.Second / elapsed)))

		// Sleep for the remainder of the frame
		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}

		elapsed = time.Since(start)
		fmt.Println("Real FPS: " + strconv.Itoa(int(time.Second / elapsed)))
	}
}

// Process an image and print it to the terminal
func processImage (img gocv.Mat) {

	// Convert the Mat to image, scaling to the terminal size
	height, width := getTerminalSize()

	// clear the screen if the terminal size has changed
	if height != terminalHeight || width != terminalWidth {
		fmt.Print("\033[2J")

		terminalHeight = height
		terminalWidth = width
	}


	switch mode {
		case "tty":
			// full blocks, we need to use double the width
			width /= 2
		case "tty_subsample":
			// full blocks subsampled, two pixels per character
			height *= 2
		case "unicode":
			// half blocks, two pixels high per character
			height *= 2
		case "truecolor":
			// truecolor, same as unicode
			height *= 2
	}

	// Add space for the fps counter
	height -= 5

	// Only resize if the image is larger than the terminal, keeping proportions
	if img.Cols() > width || img.Rows() > height {
		// check which limit is hit first
		if float64(img.Cols()) / float64(width) > float64(img.Rows()) / float64(height) {
			// width limit is hit first
			height = int(float64(img.Rows()) / float64(img.Cols()) * float64(width))
		} else {
			// height limit is hit first
			width = int(float64(img.Cols()) / float64(img.Rows()) * float64(height))
		}
	}

	gocv.Resize(img, &img, image.Point{X: width, Y: height}, 0, 0, gocv.InterpolationDefault)

	// Get image from resized Mat
	frame, err := img.ToImage()
	if err != nil {
		fmt.Println("Error converting image")
		os.Exit(1)
	}


	// Print the image
	switch mode {
		case "tty":
			// Convert the image to grayscale array
			pixels, height, width, err := importFrame(frame)
			if err != nil {
				fmt.Println("Error converting image")
				os.Exit(1)
			}
		
			// Quantize the image
			quantized := quantize(pixels, height, width, uint8(threshold))

			// Print the image using fullblocks
			printFullBlocks(quantized, height, width, 2)

		case "tty_subsample":
			// Convert the image to grayscale array
			pixels, height, width, err := importFrame(frame)
			if err != nil {
				fmt.Println("Error converting image")
				os.Exit(1)
			}

			// Subsample the image
			subsampled, height, width := subsample(pixels, height, width)

			// Quantize the image
			quantized := quantize(subsampled, height, width, uint8(threshold))

			// Print the image using fullblocks
			printFullBlocks(quantized, height, width, 1)

		case "unicode":
			// Convert the image to grayscale array
			pixels, height, width, err := importFrame(frame)
			if err != nil {
				fmt.Println("Error converting image")
				os.Exit(1)
			}

			// Quantize the image
			quantized := quantize(pixels, height, width, uint8(threshold))

			// Print the image using halfblocks
			printHalfBlocks(quantized, height, width)

		case "truecolor":
			// Convert the image to RGB array
			R, G, B, height, width, err := importFrameColor(frame)
			if err != nil {
				fmt.Println("Error converting image")
				os.Exit(1)
			}

			// Print the image using halfblocks, shading with color
			printHalfBlocksColor(R, G, B, height, width)
	}

}

// Gets the length and width of the terminal
func getTerminalSize() (height, width int) {
	width, height, err := terminal.GetSize(0)
	if err != nil {
		fmt.Println("Error getting terminal size")
		os.Exit(1)
	}

	if width < 1 || height < 1 {
		fmt.Println("Terminal size is too small: ", width, "x", height)
		os.Exit(1)
	}

	return
}

// Print out the image using black and white spaces, doubling width for perfect pixels
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

// Quantize an image array with a given threshold into a 2D array of 0s and 1s
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

// Reads an image and returns a 2D array of the pixels in grayscale
func importFrame(image image.Image) ([][]uint8, int, int, error) {

	// Get the image bounds
	bounds := image.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Create a 2D array of the pixels
	pixels := make([][]uint8, height)
	for i := range pixels {
		pixels[i] = make([]uint8, width)
	}

	// Iterate over the pixels and convert them to grayscale
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels[y][x] = color.GrayModel.Convert(image.At(x, y)).(color.Gray).Y
		}
	}

	return pixels, height, width, nil
}

// Reads an image and return 3 2D arrays of the pixels in RGB, uint32
func importFrameColor(image image.Image) ([][]uint8, [][]uint8, [][]uint8, int, int, error) {
	
	// Get the image bounds
	bounds := image.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Create a 2D array of the pixels
	r := make([][]uint8, height)
	g := make([][]uint8, height)
	b := make([][]uint8, height)
	a := make([][]uint8, height)

	for i := range r {
		r[i] = make([]uint8, width)
		g[i] = make([]uint8, width)
		b[i] = make([]uint8, width)
		a[i] = make([]uint8, width)
	}

	// Iterate over the pixels and calculate the non premultiplied values
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgba := color.RGBAModel.Convert(image.At(x, y)).(color.RGBA)
			r[y][x] = rgba.R
			g[y][x] = rgba.G
			b[y][x] = rgba.B
			a[y][x] = rgba.A
		}
	}

	return r, g, b, height, width, nil
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
func printHalfBlocksColor(r, g, b [][]uint8, height, width int) {
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
		last_top_r := r[y*2][0]
		last_bottom_r := r[y*2+1][0]

		last_top_g := g[y*2][0]
		last_bottom_g := g[y*2+1][0]

		last_top_b := b[y*2][0]
		last_bottom_b := b[y*2+1][0]

		// Loop through each column of pixels in the current row
		for x := 0; x < width; x++ {

			// Determine the color to print based on the value of the pixel
			top_r := r[y*2][x]
			bottom_r := r[y*2+1][x]

			top_g := g[y*2][x]
			bottom_g := g[y*2+1][x]

			top_b := b[y*2][x]
			bottom_b := b[y*2+1][x]

			// foreground determines the top half of the block
			if x == 0 || last_top_r != top_r || last_top_g != top_g || last_top_b != top_b {

				top_r_char := []byte(strconv.Itoa(int(top_r)))
				top_g_char := []byte(strconv.Itoa(int(top_g)))
				top_b_char := []byte(strconv.Itoa(int(top_b)))

				// Save time by not using fmt.Sprintf
				// fmt.Sprintf("\033[38;2;%d;%d;%dm", top, top, top)
				buffer = append(buffer, []byte("\033[38;2;")...)
				buffer = append(buffer, top_r_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, top_g_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, top_b_char...)
				buffer = append(buffer, []byte("m")...)

				last_top_r = top_r
				last_top_g = top_g
				last_top_b = top_b
			}

			// background determines the bottom half of the block
			if x == 0 || last_bottom_r != bottom_r || last_bottom_g != bottom_g || last_bottom_b != bottom_b {

				bottom_r_char := []byte(strconv.Itoa(int(bottom_r)))
				bottom_g_char := []byte(strconv.Itoa(int(bottom_g)))
				bottom_b_char := []byte(strconv.Itoa(int(bottom_b)))

				// Save time by not using fmt.Sprintf
				// fmt.Sprintf("\033[48;2;%d;%d;%dm", bottom, bottom, bottom)
				buffer = append(buffer, []byte("\033[48;2;")...)
				buffer = append(buffer, bottom_r_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, bottom_g_char...)
				buffer = append(buffer, []byte(";")...)
				buffer = append(buffer, bottom_b_char...)
				buffer = append(buffer, []byte("m")...)

				last_bottom_r = bottom_r
				last_bottom_g = bottom_g
				last_bottom_b = bottom_b
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

// Package i2a provides functionality to convert images to ASCII art with optional color support.
package i2a

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Decode decodes an image from the provided reader.
// Supported formats: JPEG, PNG, GIF, BMP, TIFF, and WebP.
func Decode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Config holds the configuration options for image to ASCII conversion.
type Config struct {
	// CharMap is the slice of runes used to represent different brightness levels.
	// It should be ordered from darkest to lightest.
	CharMap []rune
	// Height is the desired height of the output in characters.
	Height int
	// Width is the desired width of the output in characters.
	Width int
	// Color indicates whether colored output is enabled.
	Color bool
	// TrueColor indicates whether RGB truecolor output is enabled.
	// Only effective if Color is true.
	TrueColor bool
}

var (
	errEmptyCharMap = errors.New("charmap cannot be empty")
	errNegWidth     = errors.New("width must be greater than zero")
	errNegHeight    = errors.New("height must be greater than zero")
)

func (config Config) Validate() error {
	switch {
	case len(config.CharMap) == 0:
		return errEmptyCharMap
	case config.Width <= 0:
		return errNegWidth
	case config.Height <= 0:
		return errNegHeight
	default:
		return nil
	}
}

// ImageToASCII converts the given image to an ASCII art representation based on the provided configuration.
func ImageToASCII(img image.Image, config Config) (string, error) {
	err := config.Validate()
	if err != nil {
		return "", err
	}

	width := config.Width
	height := config.Height

	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	size := 1
	if config.Color {
		size += len([]rune(Colorize("", color.White, config.TrueColor)))
	}

	var b strings.Builder
	b.Grow(size*height*width + height)

	// TODO: resize image properly
	for y := range height {
		for x := range width {
			col := img.At(x*imgW/width, y*imgH/height)
			val := colorToChar(col, config.CharMap)
			if config.Color {
				b.WriteString(Colorize(string(val), col, config.TrueColor))
			} else {
				b.WriteRune(val)
			}
		}
		b.WriteRune('\n')
	}

	return b.String(), nil
}

func colorToChar(col color.Color, charMap []rune) rune {
	r, g, b, _ := col.RGBA()
	index := int(math.Round(float64(r+g+b) / 3 * float64(len(charMap)-1) / 0xffff))
	return charMap[index]
}

// RenderGIF renders the provided GIF image to the terminal as ASCII art based on the given configuration.
func RenderGIF(img *gif.GIF, config Config) error {
	err := config.Validate()
	if err != nil {
		return err
	}

	var (
		processed int
		wg        sync.WaitGroup
	)

	a := len(img.Image)
	numCPU := runtime.NumCPU()
	frames := make([]string, a)

	for nc := range numCPU {
		wg.Add(1)
		go func(i, n int) {
			for ; i < n; i++ {
				// TODO: handle sub-images properly
				out, _ := ImageToASCII(img.Image[i], config)
				frames[i] = out
				processed++
				fmt.Printf("\rProcessing GIF frames: %3.0f%%", float32(processed*100/a))
			}
			wg.Done()
		}(nc*a/numCPU, (nc+1)*a/numCPU)
	}
	wg.Wait()
	fmt.Print("\r                           \r")

	fmt.Print("\x1b[?25l\x1b[2J")
	defer fmt.Print("\x1b[?25h")
	for i := 0; img.LoopCount == 0 || i <= max(img.LoopCount, 0); i++ {
		for j, frame := range frames {
			fmt.Print("\x1b[H" + frame)
			time.Sleep(time.Duration(img.Delay[j]) * 10 * time.Millisecond)
		}
	}
	return nil
}

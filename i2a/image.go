// Package i2a provides functionality to convert images to ASCII art with optional color support.
package i2a

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
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
	// It should be ordered from lightest to darkest.
	CharMap []rune
	// Width is the desired width of the output in characters.
	Width int
	// Color indicates whether colored output is enabled.
	Color bool
	// TrueColor indicates whether RGB truecolor output is enabled.
	// Only effective if Color is true.
	TrueColor bool
}

type ConfigError string

func (e ConfigError) Error() string { return string(e) }

// ImageToASCII converts the given image to an ASCII art representation based on the provided configuration.
// The height is calculated to maintain the aspect ratio, considering character dimensions.
// If it returns an error, it will be of type ConfigError.
func ImageToASCII(img image.Image, config Config) (string, error) {
	if len(config.CharMap) == 0 {
		return "", ConfigError("CharMap cannot be empty")
	}

	transformFn := func(s string, _ color.Color) string { return s }
	switch {
	case config.Color && config.TrueColor:
		transformFn = ToAnsiColorString
	case config.Color:
		transformFn = ToAnsi256ColorString
	}

	width := config.Width

	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	height := width * imgH / imgW / 2

	var out string
	for y := range height {
		for x := range width {
			col := img.At(x*imgW/width, y*imgH/height)
			val := colorToChar(col, config.CharMap)
			out += transformFn(val, col)
		}
		if y != height-1 {
			out += "\n"
		}
	}
	return out, nil
}

func colorToChar(col color.Color, charMap []rune) string {
	r, g, b, _ := col.RGBA()
	index := int(math.Round(float64(r+g+b) / 3 * float64(len(charMap)-1) / 0xffff))
	return string(charMap[index])
}

// RenderGIF renders the provided GIF image to the terminal as ASCII art based on the given configuration.
func RenderGIF(img *gif.GIF, config Config) {
	frames := []string{}
	p := 100 / float64(len(img.Image))
	for _, frame := range img.Image {
		out, _ := ImageToASCII(frame, config)
		frames = append(frames, out)
		fmt.Printf("\rProcessing GIF frames: %.0f%%", float64(len(frames))*p)
	}

	fmt.Print("\x1b[2J")
	for i := 0; img.LoopCount == 0 || i <= max(img.LoopCount, 0); i++ {
		for j, frame := range frames {
			time.Sleep(time.Duration(img.Delay[j]) * 10 * time.Millisecond)
			fmt.Println("\x1b[H" + frame)
		}
	}
}

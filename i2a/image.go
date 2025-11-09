// Package i2a provides functionality to convert images to ASCII art with optional color support.
package i2a

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
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
	// Width is the desired width of the output in characters.
	Width int
	// Height is the desired height of the output in characters.
	Height int
	// Color indicates whether colored output is enabled.
	Color bool
	// TrueColor indicates whether RGB truecolor output is enabled.
	// Only effective if Color is true.
	TrueColor bool
	// Transparent indicates whether to treat transparent pixels as spaces.
	Transparent bool
	// Background indicates whether to use background colors instead of foreground colors.
	Background bool
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
		size = len(Colorize(' ', color.White, config.TrueColor, false, false))
	}

	var b strings.Builder
	b.Grow(size*height*width + height)

	// TODO: resize image properly
	for y := range height {
		for x := range width {
			col := img.At((x*imgW/width)+bounds.Min.X, (y*imgH/height)+bounds.Min.Y)
			val := colorToChar(col, config.CharMap)
			if config.Color {
				b.WriteString(Colorize(val, col, config.TrueColor, config.Transparent, config.Background))
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
// transform is a function that can be used to modify each frame's ASCII representation before rendering.
// Rendering can be aborted by sending a signal to the abort channel.
func RenderGIF(img *gif.GIF, config Config, transform func(string) string, abort <-chan struct{}) error {
	err := config.Validate()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	a := len(img.Image)
	subImages := createGIFSubImages(img)

	numCPU := runtime.NumCPU()
	frames := make([]string, a)

	wg.Add(numCPU)
	for nc := range numCPU {
		go func(i, n int) {
			for ; i < n; i++ {
				var out string
				out, _ = ImageToASCII(subImages[i], config)
				frames[i] = transform(out)
			}
			wg.Done()
		}(nc*a/numCPU, (nc+1)*a/numCPU)
	}
	wg.Wait()

	for i := 0; img.LoopCount == 0 || i <= max(img.LoopCount, 0); i++ {
		for j, frame := range frames {
			fmt.Print(frame)
			select {
			case <-abort:
				return nil
			case <-time.After(time.Duration(img.Delay[j]) * 10 * time.Millisecond):
			}
		}
	}
	return nil
}

// createGIFSubImages creates full frames for each sub-image in the GIF, taking disposal methods into account.
func createGIFSubImages(img *gif.GIF) []image.Image {
	a := len(img.Image)
	subImages := make([]image.Image, a)

	bounds := image.Rect(0, 0, img.Config.Width, img.Config.Height)
	canvas := image.NewRGBA(bounds)

	draw.Draw(canvas, bounds, img.Image[0], image.Point{}, draw.Src)

	prevState := image.NewRGBA(bounds)

	for i, frame := range img.Image {
		draw.Draw(canvas, bounds, frame, image.Point{}, draw.Over)

		currentFrame := image.NewRGBA(bounds)
		draw.Draw(currentFrame, bounds, canvas, image.Point{}, draw.Src)
		subImages[i] = currentFrame

		switch img.Disposal[i] {
		case gif.DisposalPrevious:
			draw.Draw(canvas, bounds, prevState, image.Point{}, draw.Src)
		case gif.DisposalBackground:
			draw.Draw(canvas, frame.Bounds(), image.NewUniform(color.Transparent), image.Point{}, draw.Src)
			fallthrough
		case gif.DisposalNone:
			fallthrough
		default:
			draw.Draw(prevState, bounds, canvas, image.Point{}, draw.Src)
		}
	}
	return subImages
}

package i2a

import (
	"fmt"
	"image/color"
	"math"
)

// Colorize returns the input character wrapped in an ANSI escape sequence
// that sets the color based on the provided color col.
// If trueColor is true, it uses 24-bit RGB (true color) values.
// Otherwise, it uses the 256-color ANSI palette.
// If transparent is true and the color is fully transparent, it returns a space character.
// If background is true, it sets the background color instead of the foreground color.
func Colorize(c rune, col color.Color, trueColor bool, transparent bool, background bool) string {
	r, g, b, a := col.RGBA()
	if transparent && a == 0 {
		return " "
	}

	var colorOffset byte = 38
	if background {
		colorOffset = 48
	}

	if trueColor {
		return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm%c\x1b[0m", colorOffset, r>>8, g>>8, b>>8, c)
	}
	return fmt.Sprintf("\x1b[%d;5;%dm%c\x1b[0m", colorOffset, toAnsi256(col), c)
}

func toAnsi256(c color.Color) byte {
	r, g, b, _ := c.RGBA()
	calc := func(v uint32) float64 { return math.Round(float64(v) / 0xffff * 5) }
	return byte(16 + (36 * calc(r)) + (6 * calc(g)) + calc(b))
}

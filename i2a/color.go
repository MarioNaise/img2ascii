package i2a

import (
	"fmt"
	"image/color"
	"math"
)

// Colorize returns the input string wrapped in an ANSI escape sequence
// that sets the color based on the provided color c.
// If trueColor is true, it uses 24-bit RGB (true color) values.
// Otherwise, it uses the 256-color ANSI palette.
func Colorize(s string, c color.Color, trueColor bool) string {
	r, g, b, _ := c.RGBA()
	if trueColor {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r>>8, g>>8, b>>8, s)
	}
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", toAnsi256(c), s)
}

func toAnsi256(c color.Color) byte {
	r, g, b, _ := c.RGBA()
	calc := func(v uint32) float64 { return math.Round(float64(v) / 0xffff * 5) }
	return byte(16 + (36 * calc(r)) + (6 * calc(g)) + calc(b))
}

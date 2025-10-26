package i2a

import (
	"fmt"
	"image/color"
	"math"
)

// ToAnsiColorString converts a string to an ANSI RGB truecolor escape code
func ToAnsiColorString(s string, c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r>>8, g>>8, b>>8, s)
}

// ToAnsi256ColorString converts a string to an ANSI 256 color escape code
func ToAnsi256ColorString(s string, c color.Color) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", toAnsi256(c), s)
}

func toAnsi256(c color.Color) byte {
	r, g, b, _ := c.RGBA()
	calc := func(v uint32) float64 { return math.Round(float64(v) / 0xffff * 5) }
	return byte(16 + (36 * calc(r)) + (6 * calc(g)) + calc(b))
}

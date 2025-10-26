package main

import (
	"flag"
	"os"

	"github.com/MarioNaise/img2ascii/i2a"
	"golang.org/x/term"
)

func parseFlagsToConfig() i2a.Config {
	tw := getTerminalWidth()

	charmap := flag.String("map", " .-:=+*#%@$", "Characters to use for mapping brightness levels")
	width := flag.Int("width", tw/3, "Width of the output in characters\nDefaults to one third of current stdout or 0 of it cannot be determined")
	full := flag.Bool("full", false, "Use full terminal width (overrides -width)")
	color := flag.Bool("color", false, "Enable colored output")
	truecolor := flag.Bool("truecolor", os.Getenv("COLORTERM") == "truecolor", "Use RGB truecolor for output (requires -color)\nDefaults to true if COLORTERM=truecolor is set in the environment")
	flag.Parse()

	if full != nil && *full {
		width = &tw
	}

	return i2a.Config{
		CharMap:   []rune(*charmap),
		Width:     *width,
		Color:     *color,
		TrueColor: *truecolor,
	}
}

func getTerminalWidth() int {
	if tw, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return tw
	}
	return 0
}

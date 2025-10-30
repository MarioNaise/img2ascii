package main

import (
	"flag"
	"os"

	"github.com/MarioNaise/img2ascii/i2a"
	"golang.org/x/term"
)

var cmd = flag.NewFlagSet("img2ascii", flag.ExitOnError)

type flagConfig struct {
	charMap   string
	height    int
	width     int
	color     bool
	trueColor bool
	full      bool
	animate   bool
}

func parseFlags() flagConfig {
	charmap := cmd.String("map", " .-:=+*#%@$", "Characters to use for mapping brightness levels")
	height := cmd.Int("height", 0, "Height of the output in characters\nDefaults to stdout height or 0 if it cannot be determined\n(exg. when output is redirected to a file)")
	width := cmd.Int("width", 0, "Width of the output in characters")
	full := cmd.Bool("full", false, "Use full terminal width (overrides -width and -height)")
	color := cmd.Bool("color", false, "Enable colored output")
	truecolor := cmd.Bool("truecolor", os.Getenv("COLORTERM") == "truecolor", "Use RGB truecolor for output (requires -color)\nDefaults to true if COLORTERM=truecolor is set in the environment")
	animate := cmd.Bool("animate", true, "Animate GIF images (allows only single input file)")

	cmd.Parse(os.Args[1:])

	if len(*charmap) == 0 {
		logFatal("character map cannot be empty")
	}

	if *height != 0 && *width != 0 {
		logFatal("cannot set both height and width")
	}

	return flagConfig{
		charMap:   *charmap,
		height:    *height,
		width:     *width,
		color:     *color,
		trueColor: *truecolor,
		full:      *full,
		animate:   *animate,
	}
}

func (f flagConfig) toI2AConfig(imgHeight int, imgWidth int) i2a.Config {
	var outWidth int
	var outHeight int

	tw, th := getTerminalSize()

	arTerm := float64(tw) / float64(th) / 2
	arImg := float64(imgWidth) / float64(imgHeight)

	switch {
	case arTerm < arImg && f.height == 0 && f.width == 0:
		fallthrough
	case f.full:
		outWidth = tw
		outHeight = tw * imgHeight / imgWidth / 2
	case f.width != 0:
		outWidth = f.width
		outHeight = f.width * imgHeight / imgWidth / 2
	case f.height != 0:
		outHeight = f.height
		outWidth = f.height * imgWidth * 2 / imgHeight
	default:
		outHeight = th - 1
		outWidth = (th - 1) * imgWidth * 2 / imgHeight
	}

	return i2a.Config{
		CharMap:   []rune(f.charMap),
		Width:     outWidth,
		Height:    outHeight,
		Color:     f.color,
		TrueColor: f.trueColor,
	}
}

func getTerminalSize() (int, int) {
	if tw, th, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return tw, th
	}
	return 0, 0
}

package main

import (
	"errors"
	"flag"
	"os"

	"github.com/MarioNaise/img2ascii/i2a"
	"golang.org/x/term"
)

var cmd = flag.NewFlagSet("img2ascii", flag.ExitOnError)

func usage() {
	cmd.Output().Write([]byte("Usage:\n  img2ascii [OPTIONS] [FILES...]\n\nOptions:\n"))
	cmd.PrintDefaults()
	cmd.Output().Write([]byte("  -h, -help\n\tShow this help message\n"))
}

type flagConfig struct {
	// image config
	charMap     string
	width       int
	height      int
	color       bool
	trueColor   bool
	background  bool
	transparent bool

	// cli flags
	full    string
	animate bool
}

func parseFlags() flagConfig {
	cmd.SetOutput(os.Stdout)
	cmd.Usage = usage

	charmap := cmd.String("map", " .-:=+*#%@$", "Characters to use for mapping brightness levels")
	width := cmd.Int("width", 0, "Width of the output in characters")
	height := cmd.Int("height", 0, "Height of the output in characters")
	color := cmd.Bool("color", false, "Enable colored output")
	truecolor := cmd.Bool("truecolor", os.Getenv("COLORTERM") == "truecolor",
		"Use RGB truecolor for output (requires -color)\nDefaults to true if COLORTERM=truecolor")
	transparent := cmd.Bool("transparent", false, "Treat transparent pixels as spaces")
	background := cmd.Bool("bg", false, "Use background colors instead of foreground colors (requires -color)")

	full := cmd.String("full", "",
		`Use full terminal dimensions
This is the default, if no other dimension options are provided.
Defaults to either 'w' or 'h', depending on terminal and image size.
Overrides -width and -height
Options:
 - 'w': full width
 - 'h': full height
 - 'term': full terminal size (ignoring image aspect ratio)`)
	animate := cmd.Bool("animate", true, "Animate GIF images\nAnimations can be aborted by pressing q, Esc or Space")

	cmd.Parse(os.Args[1:])

	if *full != "" && *full != "w" && *full != "h" && *full != "term" {
		logFatal("invalid value for -full; must be one of 'w', 'h', or 'term'")
	}

	f := flagConfig{
		charMap:     *charmap,
		width:       *width,
		height:      *height,
		color:       *color,
		trueColor:   *truecolor,
		transparent: *transparent,
		background:  *background,
		animate:     *animate,
		full:        *full,
	}
	c, _ := f.toI2AConfig(1, 1)
	err := c.Validate()
	if err != nil {
		logFatal(err)
	}
	return f
}

var errImgDim = errors.New("image dimensions must be greater than zero")

func (f flagConfig) toI2AConfig(imgWidth, imgHeight int) (i2a.Config, error) {
	if imgWidth <= 0 || imgHeight <= 0 {
		return i2a.Config{}, errImgDim
	}

	var outWidth int
	var outHeight int

	tw, th := getTerminalSize()
	th -= 1 // leave space for prompt

	arTerm := float64(tw) / float64(th) / 2
	arImg := float64(imgWidth) / float64(imgHeight)

	switch {
	case f.full == "term":
		outWidth = tw
		outHeight = th
	case f.full == "h":
		outHeight = th
		outWidth = th * imgWidth * 2 / imgHeight
	case f.full == "w":
		fallthrough
	case arTerm < arImg && f.height == 0 && f.width == 0:
		outWidth = tw
		outHeight = tw * imgHeight / imgWidth / 2
	case f.width != 0 && f.height != 0:
		outWidth = f.width
		outHeight = f.height
	case f.width != 0:
		outWidth = f.width
		outHeight = f.width * imgHeight / imgWidth / 2
	case f.height != 0:
		outHeight = f.height
		outWidth = f.height * imgWidth * 2 / imgHeight
	default:
		outHeight = th
		outWidth = th * imgWidth * 2 / imgHeight
	}

	return i2a.Config{
		CharMap:     []rune(f.charMap),
		Width:       outWidth,
		Height:      outHeight,
		Color:       f.color,
		TrueColor:   f.trueColor,
		Transparent: f.transparent,
		Background:  f.background,
	}, nil
}

func getTerminalSize() (int, int) {
	tw, th, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		tw, th, _ = term.GetSize(int(os.Stdin.Fd()))
	}
	return tw, th
}

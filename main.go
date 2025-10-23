package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
	"golang.org/x/term"
)

var (
	charMap   string
	width     int
	full      bool
	truecolor bool
	colored   bool
)

func main() {
	termW, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal("Error getting terminal size:", err)
	}

	flag.StringVar(&charMap, "map", " .-:=+*#%@$", "Characters to use for mapping brightness levels")
	flag.IntVar(&width, "width", termW/3, "Width of the output in characters")
	flag.BoolVar(&full, "full", false, "Use full terminal width (overrides -max-width)")
	flag.BoolVar(&colored, "color", false, "Enable colored output")
	flag.BoolVar(&truecolor, "truecolor", os.Getenv("COLORTERM") == "truecolor", "Use RGB truecolor for output (requires -color)")
	flag.Parse()

	if len(charMap) == 0 {
		log.Fatal("Error: character map cannot be empty")
	}

	if full {
		width = termW
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal("Error reading stdin fileinfo:", err)
	}

	hasArgs := len(flag.Args()) > 0
	isStdin := stat.Size() > 0

	if !hasArgs && !isStdin {
		flag.Usage()
		return
	}

	if !hasArgs && isStdin {
		processFiles([]*os.File{os.Stdin})
		return
	}

	files := []*os.File{}
	for _, arg := range flag.Args() {
		file, err := os.Open(arg)
		if err != nil {
			log.Fatal("Error opening file:", err)
		}
		files = append(files, file)
	}
	processFiles(files)
}

func processFiles(files []*os.File) {
	for _, file := range files {
		// TODO: handle gif subimages
		img, _, err := image.Decode(file)
		if err != nil {
			log.Fatal("Error decoding image:", err)
		}
		file.Close()
		processImage(img)
	}
}

func processImage(img image.Image) {
	t := func(r rune, _ color.Color) string { return string(r) }
	switch {
	case colored && truecolor:
		t = trueColorString
	case colored:
		t = colorString
	}

	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	height := width * imgH / imgW / 2

	var out string
	for y := range height {
		for x := range width {
			col := img.At(x*imgW/width, y*imgH/height)
			r, g, b, _ := col.RGBA()
			index := math.Round(float64(r+g+b) / 3 * float64(len([]rune(charMap))-1) / 0xffff)
			val := []rune(charMap)[int(index)]
			out += t(val, col)
		}
		out += "\n"
	}
	fmt.Println(out)
}

func trueColorString(char rune, c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c", r>>8, g>>8, b>>8, char)
}

func colorString(char rune, c color.Color) string {
	return fmt.Sprintf("\x1b[38;5;%dm%c", rgbToAnsi256(c), char)
}

func rgbToAnsi256(c color.Color) byte {
	r, g, b, _ := c.RGBA()
	t := func(v uint32) float64 { return math.Round(float64(v) / 0xffff * 5) }
	return byte(16 + (36 * t(r)) + (6 * t(g)) + t(b))
}

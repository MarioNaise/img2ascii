package main

import (
	"flag"
	"fmt"
	"image/gif"
	"os"
	"path/filepath"

	"github.com/MarioNaise/img2ascii/i2a"
)

func main() {
	config := parseFlagsToConfig()

	if len(config.CharMap) == 0 {
		logFatal("character map cannot be empty")
	}

	var isStdin bool
	if stat, err := os.Stdin.Stat(); err == nil {
		isStdin = stat.Size() > 0
	}

	hasArgs := flag.NArg() > 0

	if !hasArgs && !isStdin {
		flag.Usage()
		return
	}

	if !hasArgs && isStdin {
		processFiles([]*os.File{os.Stdin}, config)
		return
	}

	files := []*os.File{}
	errs := []error{}
	for _, arg := range flag.Args() {
		if filepath.Ext(arg) == ".gif" {
			if flag.NArg() > 1 {
				logFatal("can only process one gif")
			} else {
				processGIF(arg, config)
				return
			}
		}
		file, err := os.Open(arg)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to open file: %v", err))
			continue
		}
		files = append(files, file)
	}
	errs = append(errs, processFiles(files, config)...)

	logAll(errs)
}

func processGIF(path string, config i2a.Config) {
	file, err := os.Open(path)
	if err != nil {
		logFatal(fmt.Sprintf("unable to open file: %v", err))
	}
	img, err := gif.DecodeAll(file)
	if err != nil {
		logFatal(fmt.Sprintf("unable to decode gif %s: %v", file.Name(), err))
	}

	i2a.RenderGIF(img, config)
}

func processFiles(files []*os.File, config i2a.Config) []error {
	errs := []error{}
	for _, file := range files {
		defer file.Close()
		img, err := i2a.Decode(file)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to decode %s: %v", file.Name(), err))
			continue
		}
		out, _ := i2a.ImageToASCII(img, config)
		fmt.Println(out)
	}
	return errs
}

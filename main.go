package main

import (
	"fmt"
	"image/gif"
	"os"
	"path/filepath"

	"github.com/MarioNaise/img2ascii/i2a"
)

func main() {
	flags := parseFlags()

	var isStdin bool
	if stat, err := os.Stdin.Stat(); err == nil {
		isStdin = stat.Size() > 0
	}

	hasArgs := cmd.NArg() > 0

	if !hasArgs && !isStdin {
		cmd.Usage()
		return
	}

	if !hasArgs && isStdin {
		handleStdin(flags)
		return
	}

	files := []*os.File{}
	errs := []error{}
	for _, arg := range cmd.Args() {
		if filepath.Ext(arg) == ".gif" && flags.animate {
			if cmd.NArg() > 1 {
				logFatal("can only process one gif")
			} else {
				processGIF(arg, flags)
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
	errs = append(errs, processFiles(files, flags)...)

	logAll(errs)
}

func handleStdin(flags flagConfig) {
	if flags.animate {
		img, err := gif.DecodeAll(os.Stdin)
		if err != nil {
			logFatal(fmt.Sprintf("unable to decode gif: %v", err))
		}
		config := flags.toI2AConfig(img.Config.Height, img.Config.Width)
		i2a.RenderGIF(img, config)
	}
	processFiles([]*os.File{os.Stdin}, flags)
}

func processGIF(path string, flags flagConfig) {
	file, err := os.Open(path)
	if err != nil {
		logFatal(fmt.Sprintf("unable to open file: %v", err))
	}
	img, err := gif.DecodeAll(file)
	if err != nil {
		logFatal(fmt.Sprintf("unable to decode gif %s: %v", file.Name(), err))
	}
	config := flags.toI2AConfig(img.Config.Height, img.Config.Width)

	i2a.RenderGIF(img, config)
}

func processFiles(files []*os.File, flags flagConfig) []error {
	errs := []error{}
	for _, file := range files {
		defer file.Close()
		img, err := i2a.Decode(file)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to decode %s: %v", file.Name(), err))
			continue
		}
		config := flags.toI2AConfig(img.Bounds().Dy(), img.Bounds().Dx())
		out, _ := i2a.ImageToASCII(img, config)
		fmt.Println(out)
	}
	return errs
}

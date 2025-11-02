package main

import (
	"bytes"
	"fmt"
	"image/gif"
	"io"
	"os"
	"path/filepath"

	"github.com/MarioNaise/img2ascii/i2a"
)

func main() {
	flags := parseFlags()

	hasArgs := cmd.NArg() > 0

	var readStdin bool
	if stat, err := os.Stdin.Stat(); err == nil {
		readStdin = stat.Size() > 0
	}

	switch {
	case !hasArgs && !readStdin:
		cmd.Usage()
	case !hasArgs && readStdin:
		handleStdin(flags)
	default:
		handleArgs(flags)
	}
}

func handleArgs(flags flagConfig) {
	files := []*os.File{}
	errs := []error{}

	for _, arg := range cmd.Args() {
		if filepath.Ext(arg) == ".gif" && flags.animate {
			if cmd.NArg() > 1 {
				logFatal("can only animate one gif (run with --animate=0 instead)")
			} else {
				processGIF(arg, flags)
				os.Exit(0)
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
	var buf bytes.Buffer
	_, err := io.Copy(&buf, os.Stdin)
	if err != nil {
		logFatal(fmt.Sprintf("unable to read stdin: %v", err))
	}
	r := bytes.NewReader(buf.Bytes())

	if flags.animate && isGIF(buf.Bytes()) {
		img, _ := gif.DecodeAll(r)
		config, err := flags.toI2AConfig(img.Config.Width, img.Config.Height)
		if err != nil {
			logFatal(err)
		}
		i2a.RenderGIF(img, config)
		return
	}
	err = processContents(r, "stdin", flags)
	if err != nil {
		logFatal(err)
	}
}

func isGIF(data []byte) bool {
	_, err := gif.DecodeAll(bytes.NewReader(data))
	return err == nil
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
	config, err := flags.toI2AConfig(img.Config.Width, img.Config.Height)
	if err != nil {
		logFatal(fmt.Errorf("%s: %v", path, err))
	}

	i2a.RenderGIF(img, config)
}

func processFiles(files []*os.File, flags flagConfig) []error {
	errs := []error{}
	for _, file := range files {
		err := processContents(file, file.Name(), flags)
		if err != nil {
			errs = append(errs, err)
		}
		file.Close()
	}
	return errs
}

func processContents(r io.Reader, name string, flags flagConfig) error {
	img, err := i2a.Decode(r)
	if err != nil {
		return fmt.Errorf("unable to decode %s: %v", name, err)
	}
	config, err := flags.toI2AConfig(img.Bounds().Dx(), img.Bounds().Dy())
	if err != nil {
		return fmt.Errorf("%s: %v", name, err)
	}
	out, _ := i2a.ImageToASCII(img, config)
	fmt.Print(out)
	return nil
}

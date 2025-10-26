package main

import (
	"flag"
	"fmt"
	"os"

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

func processFiles(files []*os.File, config i2a.Config) []error {
	errs := []error{}
	for _, file := range files {
		defer file.Close()
		// TODO: handle gif subimages
		img, err := i2a.Decode(file)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to decode %s: %v", file.Name(), err))
			continue
		}
		// error can be ignored here, config.CharMap is not empty
		out, _ := i2a.ImageToASCII(img, config)
		fmt.Println(out)
	}
	return errs
}

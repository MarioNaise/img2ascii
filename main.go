package main

import (
	"bytes"
	"fmt"
	"image/gif"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/MarioNaise/img2ascii/i2a"
	"golang.org/x/term"
)

var abortChan = make(chan struct{}, 1)

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
		return
	case !hasArgs && readStdin:
		handleStdin(flags)
	default:
		go listenForKeyEvents(abortChan)
		handleArgs(flags)
	}

	os.Stdin.Close()
}

func listenForKeyEvents(abort chan<- struct{}) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		switch buf[0] {
		case 'q', 'Q', 27, ' ': // 'q', ESC, Space
			abort <- struct{}{}
		case 3: // Ctrl+C
			abort <- struct{}{}
			os.Exit(130)
		}
	}
}

type imageData struct {
	name string
	*bytes.Buffer
}

func handleArgs(flags flagConfig) {
	images := make([]imageData, cmd.NArg())
	errs := []error{}

	var wg sync.WaitGroup
	wg.Add(cmd.NArg())

	for i, arg := range cmd.Args() {
		go func(i int, path string) {
			defer wg.Done()
			var img imageData
			var err error

			if isLink(path) {
				img, err = fetchImageData(path)
			} else {
				img, err = readImageFile(path)
			}
			if err != nil {
				errs = append(errs, err)
				return
			}
			images[i] = img
		}(i, arg)
	}
	wg.Wait()

	for _, img := range images {
		if img.Buffer == nil {
			continue
		}

		err := processImageData(img, flags)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	logAll(errs)
}

func handleStdin(flags flagConfig) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		logFatal(fmt.Errorf("unable to read stdin: %v", err))
	}
	buf := bytes.NewBuffer(b)

	r := imageData{"stdin", buf}

	err = processImageData(r, flags)
	if err != nil {
		logFatal(err)
	}
}

func fetchImageData(path string) (imageData, error) {
	res, err := http.Get(path)
	if err != nil {
		return imageData{}, fmt.Errorf("unable to download %s:\n\t%v", path, err)
	}
	buf, err := readAndClose(res.Body)
	if err != nil {
		return imageData{}, fmt.Errorf("unable to read response body from %s:\n\t%v", path, err)
	}

	return imageData{path, buf}, nil
}

func readImageFile(path string) (imageData, error) {
	file, err := os.Open(path)
	if err != nil {
		return imageData{}, fmt.Errorf("unable to open file: %v", err)
	}
	buf, err := readAndClose(file)
	if err != nil {
		return imageData{}, fmt.Errorf("unable to read file %s:\t%v", path, err)
	}
	return imageData{file.Name(), buf}, nil
}

func processImageData(img imageData, flags flagConfig) (err error) {
	if flags.animate && isGIF(img.Bytes()) {
		err = renderGIF(img, flags)
	} else {
		err = printImage(img, flags)
	}
	return err
}

func renderGIF(i imageData, flags flagConfig) error {
	img, err := gif.DecodeAll(bytes.NewReader(i.Bytes()))
	if err != nil {
		return fmt.Errorf("unable to decode gif %s: %v", i.name, err)
	}

	config, err := flags.toI2AConfig(img.Config.Width, img.Config.Height)
	if err != nil {
		return fmt.Errorf("%s: %v", i.name, err)
	}

	fd := int(os.Stdin.Fd())
	termState, makeRawErr := term.MakeRaw(fd)
	if makeRawErr == nil {
		defer term.Restore(fd, termState)
	}

	t := func(s string) string {
		if makeRawErr == nil {
			s = strings.ReplaceAll(s, "\n", "\r\n")
		}
		return fmt.Sprintf(
			"%s\x1b[%dA\r",
			s,
			config.Height,
		)
	}

	fmt.Print("\x1b[?25l")
	defer fmt.Printf("\x1b[%dB\x1b[?25h", config.Height)

	return i2a.RenderGIF(img, config, t, abortChan)
}

func printImage(i imageData, flags flagConfig) error {
	img, err := i2a.Decode(bytes.NewReader(i.Bytes()))
	if err != nil {
		return fmt.Errorf("unable to decode %s: %v", i.name, err)
	}
	config, err := flags.toI2AConfig(img.Bounds().Dx(), img.Bounds().Dy())
	if err != nil {
		return fmt.Errorf("%s: %v", i.name, err)
	}
	out, _ := i2a.ImageToASCII(img, config)
	fmt.Print(out)
	return nil
}

func readAndClose(r io.ReadCloser) (*bytes.Buffer, error) {
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func isLink(path string) bool {
	_, err := url.ParseRequestURI(path)
	return err == nil
}

func isGIF(b []byte) bool {
	_, err := gif.DecodeAll(bytes.NewReader(b))
	return err == nil
}

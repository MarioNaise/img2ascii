package main

import (
	"log"
	"os"
)

var errorLog = log.New(os.Stderr, "Error: ", 0)

func logFatal(v ...any) {
	errorLog.Fatalln(v...)
}

func logln(v ...any) {
	errorLog.Println(v...)
}

func logAll(errs []error) {
	for _, err := range errs {
		logln(err)
	}
}

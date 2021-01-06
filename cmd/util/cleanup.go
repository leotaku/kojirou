package util

import (
	"fmt"
	"os"
	"os/signal"
)

var clean = make(chan func(), 100)

func Cleanup(f func()) {
	clean <- f
}

func InitCleanup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			RunCleanup()
			fmt.Println(sig)
			os.Exit(2)
		}
	}()
}

func RunCleanup() {
	// Reverse order
	fs := make([]func(), 0)
	close(clean)
	for f := range clean {
		fs = append(fs, f)
	}
	// Cleanup
	for _, f := range fs {
		f()
	}
}

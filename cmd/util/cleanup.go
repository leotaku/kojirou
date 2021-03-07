package util

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
)

var clean = make(chan func(), 1000)

func Cleanup(f func()) {
	clean <- f
}

func InitCleanup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			RunCleanup()
			fmt.Fprintln(os.Stderr, sig)
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

func SetupDirectories(dirs ...string) error {
	for _, dir := range dirs {
		cleanupDirectory(dir)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanupDirectory(dir string) {
	for ; dir != "." && dir != "/"; dir = path.Dir(dir) {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			tmp := dir
			Cleanup(func() { os.Remove(tmp) })
		}
	}
}

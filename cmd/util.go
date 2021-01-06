package cmd

import (
	"encoding/json"
	"fmt"
	"image/jpeg"
	"os"
	"path"

	"github.com/leotaku/manki/cmd/util"
	"github.com/leotaku/mobi"
)

func setupDirectories(dirs ...string) error {
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
			util.Cleanup(func() { os.Remove(tmp) })
		}
	}
}

func writeBook(book mobi.Book, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	db := book.Realize()
	err = db.Write(f)
	if err != nil {
		return err
	}

	return nil
}

func writeThumb(book mobi.Book, dir string) error {
	if book.CoverImage != nil {
		path := path.Join(dir, book.GetThumbFilename())
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		err = jpeg.Encode(f, book.CoverImage, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func jsonprint(v interface{}) {
	json, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(json))
}

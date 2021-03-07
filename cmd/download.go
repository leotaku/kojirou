package cmd

import (
	"fmt"
	"image/jpeg"
	"os"
	"path"

	"github.com/leotaku/manki/cmd/util"
	"github.com/leotaku/manki/mangadex"
	"github.com/leotaku/mobi"
)

func downloadManga(id int) (*mangadex.Manga, error) {
	manga, err := util.Client.FetchManga(id)
	if err != nil {
		return nil, err
	}

	chs, err := util.Client.FetchChapters(id)
	if err != nil {
		return nil, err
	}

	cos, err := util.Client.FetchCovers(id)
	if err != nil {
		return nil, err
	}

	lang := util.MatchLang(langArg)
	chapters, err := filter(chs, lang)
	if err != nil {
		return nil, err
	}

	pb := util.NewBar().Message("Covers")
	covers, err := util.FetchCovers(cos, pb)
	if err != nil {
		return nil, err
	}
	pb.Finish()

	incomplete := manga.WithChapters(*chapters).WithCovers(covers)
	return &incomplete, nil
}

func downloadWriteVolumes(m mangadex.Manga, root string, thumbRoot *string) error {
	for _, idx := range m.Keys() {
		// Variables
		path := fmt.Sprintf("%v/%v.azw3", root, idx)
		pb := util.NewBar().Message(fmt.Sprintf("Volume %v", idx))
		pb.AddTotal(1)

		// Abort if file exists
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			pb.Succeed("File exists").Finish()
			continue
		}

		// Fetch volume images
		filtered := m.Chapters().FilterBy(func(ci mangadex.ChapterInfo) bool {
			return ci.VolumeIdentifier == idx
		})
		pages, err := util.FetchChapters(filtered, pb)
		if err != nil {
			return err
		}

		// Write book and thumbnail
		manga := m.WithChapters(filtered).WithPages(pages)
		mobi := util.VolumesToMobi(manga)
		err = writeBook(mobi, path)
		if err != nil {
			util.Cleanup(func() { os.Remove(path) })
			return err
		}

		if thumbRoot != nil {
			err := writeThumb(mobi, *thumbRoot)
			if err != nil {
				return err
			}
		}

		// Done
		pb.Increment()
		pb.Finish()
	}

	return nil
}

func writeBook(book mobi.Book, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	db := book.Realize()
	err = db.Write(f)
	if err != nil {
		return err
	}

	return nil
}

func writeThumb(book mobi.Book, root string) error {
	if book.CoverImage != nil {
		path := path.Join(root, book.GetThumbFilename())
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		err = jpeg.Encode(f, book.CoverImage, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

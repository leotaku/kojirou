package cmd

import (
	"fmt"
	"os"

	"github.com/leotaku/manki/cmd/util"
	"github.com/leotaku/manki/mangadex"
)

type downloadInfo struct {
	incomplete mangadex.Manga
	chapters   mangadex.ChapterList
	covers     []mangadex.ImageInfo
}

func preDownload(id int) (*downloadInfo, error) {
	b, ch, co, err := util.RetryFetch(id)
	if err != nil {
		return nil, err
	}

	lang := util.MatchLang(langArg)
	chapters, err := filter(ch, lang)
	if err != nil {
		return nil, err
	}

	incomplete := mangadex.Rebuild(*b, *chapters)
	pb := util.NewBar(fmt.Sprintf("Covers"))
	covers, err := util.FetchCovers(co, pb)
	if err != nil {
		return nil, err
	}
	pb.Finish()

	return &downloadInfo{
		incomplete: incomplete,
		chapters:   *chapters,
		covers:     covers,
	}, nil
}

func download(do downloadInfo, root string, thumbRoot *string) error {
	for _, idx := range do.incomplete.Sorted() {
		// Variables
		path := fmt.Sprintf("%v/%v.azw3", root, idx)
		pb := util.NewBar(fmt.Sprintf("Volume %v", idx))
		pb.AddTotal(1)

		// Fetch volume images
		filtered := do.chapters.FilterBy(func(ci mangadex.ChapterInfo) bool {
			return ci.VolumeIdentifier == idx
		})
		pages, err := util.FetchChapters(filtered, pb)
		if err != nil {
			return err
		}

		// Write book and thumbnail
		if _, err := os.Stat(path); os.IsNotExist(err) {
			manga := do.incomplete.WithPages(pages).WithCovers(do.covers)
			mobi := util.VolumeToMobi(manga.Info, manga.Volumes[idx])
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
		}

		// Done
		pb.Increment()
		pb.Finish()
	}

	return nil
}

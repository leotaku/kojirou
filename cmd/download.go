package cmd

import (
	"fmt"
	"image/jpeg"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/leotaku/kojirou/cmd/util"
	"github.com/leotaku/kojirou/mangadex"
	"github.com/leotaku/mobi"
)

func downloadMetaFor(id string, filter Filter) (*mangadex.Manga, error) {
	manga, err := util.Client.FetchManga(id)
	if err != nil {
		return nil, err
	}

	chs, err := util.Client.FetchChapters(id)
	if err != nil {
		return nil, err
	}

	filtered, err := filter(chs)
	if err != nil {
		return nil, err
	} else if len(filtered) == 0 {
		return nil, fmt.Errorf("no matching scantlations found")
	}

	simpleColorPrint("Title: ", manga.Info.Title, ", Authors: ", manga.Info.Authors)
	printGroupMapping(filtered)

	result := manga.WithChapters(filtered)
	return &result, nil
}

func downloadAndWrite(m mangadex.Manga, root string, thumbRoot *string, force bool) error {
	for _, idx := range m.Keys() {
		// Variables
		path := fmt.Sprintf("%v/%v.azw3", root, idx)
		pb := util.NewBar().Message(fmt.Sprintf("Volume %v", idx))
		pb.AddTotal(1)

		// Abort if file exists
		if _, err := os.Stat(path); !force && !os.IsNotExist(err) {
			pb.Succeed("File exists").Finish()
			continue
		}

		// Fetch volume images
		filtered := m.Chapters().FilterBy(func(ci mangadex.ChapterInfo) bool {
			return ci.VolumeIdentifier == idx
		})
		pages, err := util.FetchPages(filtered, pb)
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

var groupColors = []*color.Color{
	color.New(color.FgRed),
	color.New(color.FgBlue),
	color.New(color.FgMagenta),
	color.New(color.FgCyan),
	color.New(color.FgGreen),
	color.New(color.FgYellow),
	color.New(color.ReverseVideo, color.FgRed),
	color.New(color.ReverseVideo, color.FgBlue),
	color.New(color.ReverseVideo, color.FgMagenta),
	color.New(color.ReverseVideo, color.FgCyan),
	color.New(color.ReverseVideo, color.FgGreen),
	color.New(color.ReverseVideo, color.FgYellow),
}

func printGroupMapping(cl mangadex.ChapterList) {
	groupMapping := make(map[string]int)
	ids := make([]string, 0)
	for _, ci := range cl {
		idx := len(groupMapping) - 1
		if val, ok := groupMapping[gid(ci)]; ok {
			idx = val
		} else {
			idx += 1
			groupMapping[gid(ci)] = idx
		}
		ids = append(ids, groupColors[idx%len(groupColors)].Sprint(ci.Identifier))
	}

	groups := make([]string, len(groupMapping))
	for key, val := range groupMapping {
		groups[val] = groupColors[val%len(groupColors)].Sprint(key)
	}

	fmt.Printf("Groups: %v\n", strings.Join(groups, ", "))
	fmt.Printf("Chapters: %v\n", strings.Join(ids, ", "))
}

func simpleColorPrint(ss ...interface{}) {
	for n := 0; n < len(ss); n += 2 {
		underlined := color.New(color.Underline).Sprint(ss[n+1])
		fmt.Printf("%v%v", ss[n], underlined)
	}
	fmt.Println()
}

func writeBook(book mobi.Book, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return book.Realize().Write(f)
}

func writeThumb(book mobi.Book, root string) error {
	if book.CoverImage != nil {
		path := path.Join(root, book.GetThumbFilename())
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		return jpeg.Encode(f, book.CoverImage, nil)
	}

	return nil
}

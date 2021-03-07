package cmd

import (
	"fmt"
	"image/jpeg"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/leotaku/manki/cmd/util"
	"github.com/leotaku/manki/mangadex"
	"github.com/leotaku/mobi"
)

func downloadMetaFor(id int, filter Filter) (*mangadex.Manga, error) {
	manga, err := util.Client.FetchManga(id)
	if err != nil {
		return nil, err
	}

	chs, err := util.Client.FetchChapters(id)
	if err != nil {
		return nil, err
	}

	filtered := filter(chs)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no matching scantlations found")
	}

	authors := strings.Join(manga.Info.Authors, " and ")
	simpleColorPrint("Title: ", manga.Info.Title, ", Authors: ", authors)
	printGroupMapping(filtered)

	result := manga.WithChapters(filtered)
	return &result, nil
}

func downloadAddCovers(m mangadex.Manga) (*mangadex.Manga, error) {
	cos, err := util.Client.FetchCovers(m.Info.ID)
	if err != nil {
		return nil, err
	}
	pb := util.NewBar().Message("Covers")
	covers, err := util.FetchCovers(cos, pb)
	if err != nil {
		return nil, err
	}
	pb.Finish()

	result := m.WithCovers(covers)
	return &result, nil
}

func downloadAndWrite(ma mangadex.Manga, root string, thumbRoot *string) error {
	m, err := downloadAddCovers(ma)
	if err != nil {
		return err
	}

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

var groupColors = []color.Attribute{
	color.FgRed,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
	color.FgGreen,
	color.FgYellow,
	color.BgRed,
	color.BgBlue,
	color.BgMagenta,
	color.BgCyan,
	color.BgGreen,
}

func printGroupMapping(cl mangadex.ChapterList) {
	groupMapping := make(map[string]int)
	ids := make([]string, 0)
	for _, ci := range cl {
		attr := len(groupMapping) - 1
		if val, ok := groupMapping[gid(ci)]; ok {
			attr = val
		} else {
			attr += 1
			groupMapping[gid(ci)] = attr
		}
		ids = append(ids, color.New(groupColors[attr%len(groupColors)]).Sprint(ci.Identifier))
	}

	groups := make([]string, len(groupMapping))
	for key, val := range groupMapping {
		groups[val] = color.New(groupColors[val%len(groupColors)]).Sprint(key)
	}

	fmt.Printf("Groups: %v\n", strings.Join(groups, ", "))
	fmt.Printf("Chapters: %v\n", strings.Join(ids, ", "))
}

func simpleColorPrint(ss ...string) {
	for n := 0; n < len(ss); n += 2 {
		fmt.Print(ss[n])
		color.New(color.Underline).Print(ss[n+1])
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

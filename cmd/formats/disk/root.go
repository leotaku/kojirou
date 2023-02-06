package disk

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

func LoadSkeleton(directory string) (*md.Manga, error) {
	info := md.MangaInfo{
		Title: path.Base(directory),
	}

	return &md.Manga{
		Info:    info,
		Volumes: make(map[md.Identifier]md.Volume, 0),
	}, nil
}

func LoadChapters(directory string, lang language.Tag, p formats.Progress) (md.ChapterList, error) {
	result := make(md.ChapterList, 0)
	volumes, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("list '%v': %w", directory, err)
	}
	for _, volume := range volumes {
		if !volume.IsDir() {
			continue
		}
		chapters, err := os.ReadDir(path.Join(directory, volume.Name()))
		if err != nil {
			return nil, fmt.Errorf("list '%v': %w", directory, err)
		}
		for _, chapter := range chapters {
			if !chapter.IsDir() {
				continue
			}
			p.Increase(1)
			p.Add(1)

			info := md.ChapterInfo{
				Identifier:       md.NewIdentifier(chapter.Name()),
				VolumeIdentifier: md.NewIdentifier(volume.Name()),
				GroupNames:       []string{"Filesystem"},
				Language:         lang,
				ID:               path.Join(directory, volume.Name(), chapter.Name()),
			}
			result = append(result, md.Chapter{
				Info:  info,
				Pages: make(map[int]image.Image, 0),
			})
		}
	}

	return result, nil
}

func LoadPages(cl md.ChapterList, p formats.Progress) (md.ImageList, error) {
	result := make(md.ImageList, 0)
	for _, chap := range cl {
		pages, err := os.ReadDir(chap.Info.ID)
		if err != nil {
			return nil, fmt.Errorf("list '%v': %w", chap.Info.Identifier, err)
		}

		p.Increase(len(pages))
		for _, page := range pages {
			p.Add(1)

			id, err := strconv.Atoi(strings.TrimSuffix(page.Name(), path.Ext(page.Name())))
			if err != nil {
				continue
			}

			f, err := os.Open(path.Join(chap.Info.ID, page.Name()))
			if err != nil {
				return nil, err
			}
			img, _, err := image.Decode(f)
			if err != nil {
				return nil, err
			}

			result = append(result, md.Image{
				Image:             img,
				ImageIdentifier:   id,
				ChapterIdentifier: chap.Info.Identifier,
				VolumeIdentifier:  chap.Info.VolumeIdentifier,
			})
		}
	}

	return result, nil
}

func LoadCovers(directory string, p formats.Progress) (md.ImageList, error) {
	result := make(md.ImageList, 0)
	volumes, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("list '%v': %w", directory, err)
	}
	p.Increase(len(volumes))
	for _, volume := range volumes {
		if !volume.IsDir() {
			continue
		}

		img, err := readImage(directory, volume.Name())
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("image: %w", err)
		}
		result = append(result, md.Image{
			Image:            img,
			VolumeIdentifier: md.NewIdentifier(volume.Name()),
		})
	}

	return result, nil
}

func readImage(directory, name string) (image.Image, error) {
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif"} {
		f, err := os.Open(path.Join(directory, name+ext))
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("open: %w", err)
		} else {
			img, _, err := image.Decode(f)
			if err != nil {
				return nil, fmt.Errorf("decode: %w", err)
			} else {
				return img, nil
			}
		}
	}

	return nil, fmt.Errorf("no image found")
}

package mangadex

import (
	"fmt"
	"image"

	"github.com/leotaku/manki/mangadex/api"
	"golang.org/x/text/language"
)

type ImageInfo struct {
	Image image.Image

	// hidden
	imageId   int
	chapterId Identifier
	volumeId  Identifier
}

type PathInfo struct {
	Url string

	// hidden
	imageId   int
	chapterId Identifier
	volumeId  Identifier
}

type ChapterInfo struct {
	Title      string
	Views      int
	Region     language.Region
	GroupNames []string
	Hash       string
	Id         int

	// hidden
	Identifier       Identifier
	VolumeIdentifier Identifier
}

type MangaInfo struct {
	Title    string
	Authors  []string
	Artists  []string
	IsHentai bool
	Id       int
}

func Fetch(mangaID int) (*MangaInfo, ChapterList, []PathInfo, error) {
	b, err := FetchBase(mangaID)
	if err != nil {
		return nil, nil, nil, err
	}

	ch, err := FetchChapters(mangaID)
	if err != nil {
		return nil, nil, nil, err
	}

	co, err := FetchCovers(mangaID)
	if err != nil {
		return nil, nil, nil, err
	}

	return b, ch, co, nil
}

func FetchBase(mangaID int) (*MangaInfo, error) {
	b, err := api.FetchBase(mangaID)
	if err != nil {
		return nil, fmt.Errorf("Fetch covers: %w", err)
	}

	base := convertBase(b.Data)
	return &base, nil
}

func FetchChapters(mangaID int) (ChapterList, error) {
	ca, err := api.FetchChapters(mangaID)
	if err != nil {
		return nil, fmt.Errorf("Fetch chapters: %w", err)
	}

	chapters := convertChapters(ca.Data)
	return chapters, nil
}

func FetchCovers(mangaID int) ([]PathInfo, error) {
	co, err := api.FetchCovers(mangaID)
	if err != nil {
		return nil, fmt.Errorf("Fetch covers: %w", err)
	}

	covers := convertCovers(co.Data)
	return covers, nil
}

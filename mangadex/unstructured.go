package mangadex

import (
	"image"
	"time"

	"golang.org/x/text/language"
)

type ImageItem struct {
	Image image.Image

	// identifiers
	ImageIdentifier   int
	ChapterIdentifier Identifier
	VolumeIdentifier  Identifier
}

type PathItem struct {
	URL string

	// identifiers
	ImageIdentifier   int
	ChapterIdentifier Identifier
	VolumeIdentifier  Identifier
}

func (i PathItem) WithImage(img image.Image) ImageItem {
	return ImageItem{
		Image:             img,
		ChapterIdentifier: i.ChapterIdentifier,
		VolumeIdentifier:  i.VolumeIdentifier,
		ImageIdentifier:   i.ImageIdentifier,
	}
}

type ChapterInfo struct {
	Title      string
	Views      int
	Language   language.Tag
	GroupNames multiple
	PagePaths  []string
	Published  time.Time
	Hash       string
	ID         string

	// identifiers
	Identifier       Identifier
	VolumeIdentifier Identifier
}

type MangaInfo struct {
	Title   string
	Authors multiple
	Artists multiple
	ID      string
}

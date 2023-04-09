package mangadex

import (
	"image"
	"time"

	"golang.org/x/text/language"
)

type MangaInfo struct {
	Title   string
	Authors multiple
	Artists multiple
	ID      string
}

type VolumeInfo struct {
	Identifier Identifier
}

type ChapterInfo struct {
	Title      string
	Views      int
	Language   language.Tag
	GroupNames multiple
	Published  time.Time
	ID         string

	// identifiers
	Identifier       Identifier
	VolumeIdentifier Identifier
}

type Image struct {
	Image image.Image

	// identifiers
	ImageIdentifier   int
	ChapterIdentifier Identifier
	VolumeIdentifier  Identifier
}

type Path struct {
	DataURL      string
	DataSaverURL string

	// identifiers
	ImageIdentifier   int
	ChapterIdentifier Identifier
	VolumeIdentifier  Identifier
}

func (i Path) WithImage(img image.Image) Image {
	return Image{
		Image:             img,
		ChapterIdentifier: i.ChapterIdentifier,
		VolumeIdentifier:  i.VolumeIdentifier,
		ImageIdentifier:   i.ImageIdentifier,
	}
}

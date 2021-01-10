package mangadex

import (
	"image"

	"golang.org/x/text/language"
)

type ImageItem struct {
	Image image.Image

	// hidden
	imageId   int
	chapterId Identifier
	volumeId  Identifier
}

type PathItem struct {
	Url string

	// hidden
	imageId   int
	chapterId Identifier
	volumeId  Identifier
}

func (i PathItem) WithImage(img image.Image) ImageItem {
	return ImageItem{
		Image:     img,
		chapterId: i.chapterId,
		volumeId:  i.volumeId,
		imageId:   i.imageId,
	}
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

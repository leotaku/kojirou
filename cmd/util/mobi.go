package util

import (
	"fmt"
	"image"
	"math/rand"
	"time"

	"github.com/leotaku/manki/mangadex"
	"github.com/leotaku/mobi"
	"github.com/leotaku/mobi/records"
	"golang.org/x/text/language"
)

const htmlTag = `<div style="display: none">.</div><img src="kindle:embed:%v?mime=image/jpeg">`

const baseCSS = `
.image {
    display: block;
    vertical-align: baseline;
    margin: 0;
    padding: 0;
}
`

func VolumeToMobi(manga mangadex.MangaInfo, vol mangadex.Volume) mobi.Book {
	chapters := make([]mobi.Chapter, 0)
	images := make([]image.Image, 0)
	id := 0

	// Pages
	groupNames := make([]string, 0)
	for _, idx := range vol.Keys() {
		chap := vol.Chapters[idx]
		groupNames = append(groupNames, chap.Info.GroupNames...)
		links := make([]string, 0)
		for _, img := range chap.Sorted() {
			images = append(images, img)
			links = append(links, fmt.Sprintf(htmlTag, records.To32(id)))
			id++
		}
		title := chap.Info.Title
		if len(title) == 0 {
			title = fmt.Sprintf("Chapter %v", chap.Info.Identifier)
		}
		chapters = append(chapters, mobi.Chapter{
			Title:  title,
			Chunks: mobi.SingleChunks(links...),
		})
	}

	// Variables
	title := fmt.Sprintf("%v: %v", manga.Title, vol.Identifier)
	lang, _ := language.Compose()
	match, _, _ := matcher.Match(lang)

	return mobi.Book{
		Title:        title,
		Authors:      manga.Authors,
		Contributors: groupNames,
		CreatedDate:  time.Now(),
		Language:     match,
		FixedLayout:  true,
		CoverImage:   vol.Cover,
		Images:       images,
		Chapters:     chapters,
		CSSFlows:     []string{baseCSS},
		UniqueID:     rand.Uint32(),
	}
}

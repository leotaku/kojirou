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

func VolumesToMobi(manga mangadex.Manga) mobi.Book {
	chapters := make([]mobi.Chapter, 0)
	images := make([]image.Image, 0)
	id := 1

	// Pages
	groupNames := make([]string, 0)
	for _, vol := range manga.Sorted() {
		for _, chap := range vol.Sorted() {
			groupNames = unifyStrings(groupNames, chap.Info.GroupNames...)
			pages := make([]string, 0)
			for _, img := range chap.Sorted() {
				images = append(images, img)
				pages = append(pages, fmt.Sprintf(htmlTag, records.To32(id)))
				id++
			}
			title := chap.Info.Title
			if len(title) == 0 {
				title = fmt.Sprintf("Chapter %v", chap.Info.Identifier)
			}
			chapters = append(chapters, mobi.Chapter{
				Title:  title,
				Chunks: mobi.Chunks(pages...),
			})
		}
	}

	return mobi.Book{
		Title:        mangaToTitle(manga),
		Authors:      manga.Info.Authors,
		Contributors: groupNames,
		CreatedDate:  time.Now(),
		Language:     mangaToLanguage(manga),
		FixedLayout:  true,
		CoverImage:   mangaToCover(manga),
		Images:       images,
		Chapters:     chapters,
		CSSFlows:     []string{basePageCSS},
		UniqueID:     rand.Uint32(),
	}
}

func mangaToTitle(manga mangadex.Manga) string {
	nums := make([]string, 0)
	for _, idx := range manga.Keys() {
		nums = append(nums, idx.String())
	}
	sn := strings.Join(nums, ", ")

	return fmt.Sprintf("%v: %v", manga.Info.Title, sn)
}

func mangaToCover(manga mangadex.Manga) image.Image {
	return manga.Sorted()[0].Cover
}

func mangaToLanguage(manga mangadex.Manga) language.Tag {
	region := language.Region{}
	for _, chap := range manga.Chapters() {
		if region.Contains(chap.Region) {
			region = chap.Region
		} else {
			panic("unsupported: multiple different languages")
		}
	}

	lang, _ := language.Compose(region)
	match, _, _ := matcher.Match(lang)

	return match
}

func unifyStrings(this []string, other ...string) []string {
	mapping := make(map[string]struct{})
	for _, s := range this {
		mapping[s] = struct{}{}
	}
	for _, s := range other {
		mapping[s] = struct{}{}
	}

	result := make([]string, 0)
	for s := range mapping {
		result = append(result, s)
	}

	return result
}

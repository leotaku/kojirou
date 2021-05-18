package util

import (
	"fmt"
	"hash/fnv"
	"html/template"
	"image"
	"strings"
	"time"

	"github.com/leotaku/kojirou/mangadex"
	"github.com/leotaku/mobi"
	"github.com/leotaku/mobi/records"
	"golang.org/x/text/language"
)

const (
	pageTemplateString = `<div>.</div><img src="kindle:embed:{{ . }}?mime=image/jpeg">`
	basePageCSS        = `
div {
    display: none
}

img {
    display: block;
    vertical-align: baseline;
    margin: 0;
    padding: 0;
}`
)

var pageTemplate = template.Must(template.New("page").Parse(pageTemplateString))

func VolumesToMobi(manga mangadex.Manga) mobi.Book {
	chapters := make([]mobi.Chapter, 0)
	images := make([]image.Image, 0)
	pageID := 1

	// Pages
	groupNames := make([]string, 0)
	for _, vol := range manga.Sorted() {
		for _, chap := range vol.Sorted() {
			groupNames = unifyStrings(groupNames, chap.Info.GroupNames...)
			pages := make([]string, 0)
			for _, img := range chap.Sorted() {
				images = append(images, img)
				pages = append(pages, executeTemplate(pageTemplate, records.To32(pageID)))
				pageID++
			}
			title := fmt.Sprintf("%v: %v", chap.Info.Identifier, chap.Info.Title)
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
		UniqueID:     mangaToUniqueID(manga),
	}
}

func mangaToUniqueID(manga mangadex.Manga) uint32 {
	hash := fnv.New32()
	hash.Write([]byte(manga.Info.ID))
	for _, idx := range manga.Keys() {
		hash.Write([]byte(idx.String()))
	}

	return hash.Sum32()
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
	chaps := manga.Chapters()
	if len(chaps) != 0 {
		return chaps[0].Language
	} else {
		panic("unsupported: multiple different languages")
	}
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

func executeTemplate(tpl *template.Template, data interface{}) string {
	b := new(strings.Builder)
	if err := tpl.Execute(b, data); err != nil {
		panic(err)
	}

	return b.String()
}

package kindle

import (
	"fmt"
	"hash/fnv"
	"html/template"
	"image"
	"sort"
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

func GenerateMOBI(manga mangadex.Manga) mobi.Book {
	chapters := make([]mobi.Chapter, 0)
	images := make([]image.Image, 0)
	pageImageIndex := 1

	groupNames := make([]string, 0)
	for _, vol := range manga.Sorted() {
		for _, chap := range vol.Sorted() {
			groupNames = append(groupNames, chap.Info.GroupNames...)
			pages := make([]string, 0)
			for _, img := range chap.Sorted() {
				images = append(images, img)
				pages = append(pages, templateToString(pageTemplate, records.To32(pageImageIndex)))
				pageImageIndex++
			}
			title := fmt.Sprintf("%v: %v", chap.Info.Identifier, chap.Info.Title)
			chapters = append(chapters, mobi.Chapter{
				Title:  title,
				Chunks: mobi.Chunks(pages...),
			})
		}
	}
	groupNames = deduplicate(groupNames)

	return mobi.Book{
		Title:        mangaToTitle(manga),
		Authors:      manga.Info.Authors,
		Contributors: groupNames,
		CreatedDate:  time.Unix(0, 0),
		Language:     mangaToLanguage(manga),
		FixedLayout:  true,
		RightToLeft:  true,
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
	if len(chaps) == 0 {
		return language.Und
	} else {
		matcher := language.NewMatcher(mobi.SupportedLocales)
		// multiple languages are not supported
		_, i, _ := matcher.Match(chaps[0].Info.Language)
		return mobi.SupportedLocales[i]
	}
}

func deduplicate(slice []string) []string {
	sort.Stable(sort.StringSlice(slice))
	dedup := make([]string, 0)

	for i, it := range slice {
		if len(dedup) == 0 || slice[i-1] != it {
			dedup = append(dedup, it)
		}
	}
	return dedup
}

func templateToString(tpl *template.Template, data interface{}) string {
	buf := new(strings.Builder)
	if err := tpl.Execute(buf, data); err != nil {
		panic(err)
	}

	return buf.String()
}

package mangadex

import (
	"image"
	"reflect"
	"strings"

	"github.com/leotaku/kojirou/mangadex/api"
	"golang.org/x/text/language"
)

func convertManga(b *api.Manga, authors, artists *api.AuthorList) MangaInfo {
	authorNames := make([]string, 0)
	for _, a := range authors.Results {
		authorNames = append(authorNames, a.Data.Attributes.Name)
	}

	artistNames := make([]string, 0)
	for _, a := range artists.Results {
		artistNames = append(artistNames, a.Data.Attributes.Name)
	}

	return MangaInfo{
		Title:   first(b.Data.Attributes.Title),
		Authors: authorNames,
		Artists: artistNames,
		ID:      b.Data.ID,
	}
}

func convertChapters(ca []api.Chapter, groupMap map[string]api.Group) ChapterList {
	sorted := make(ChapterList, 0)
	for _, info := range ca {
		lang, _ := language.Parse(info.Data.Attributes.TranslatedLanguage)
		groups := make([]string, 0)
		for _, id := range info.Relationships.Group {
			groups = append(groups, groupMap[id].Data.Attributes.Name)
		}

		sorted = append(sorted, Chapter{
			Info: ChapterInfo{
				Title:            info.Data.Attributes.Title,
				Language:         lang,
				Views:            0, // FIXME
				Hash:             info.Data.Attributes.Hash,
				GroupNames:       groups,
				Published:        info.Data.Attributes.PublishAt,
				ID:               info.Data.ID,
				Identifier:       NewWithFallback(info.Data.Attributes.Chapter, info.Data.Attributes.Title),
				VolumeIdentifier: NewWithFallback(info.Data.Attributes.Volume, "Special"),
			},
			PagePaths: info.Data.Attributes.Data,
			Pages:     make(map[int]image.Image),
		})
	}

	reverse(sorted)
	return sorted
}

func convertCovers(coverBaseURL string, mangaID string, co []api.Cover) PathList {
	result := make(PathList, 0)
	for _, info := range co {
		url := strings.Join([]string{coverBaseURL, mangaID, info.Data.Attributes.FileName}, "/")
		result = append(result, Path{
			URL:               url,
			ImageIdentifier:   0,
			ChapterIdentifier: NewIdentifier("0"),
			VolumeIdentifier:  NewWithFallback(info.Data.Attributes.Volume, "Special"),
		})
	}

	return result
}

func convertChapter(baseURL string, ch *Chapter) PathList {
	result := make(PathList, 0)
	for i, filename := range ch.PagePaths {
		url := strings.Join([]string{baseURL, "data", ch.Info.Hash, filename}, "/")
		result = append(result, Path{
			URL:               url,
			ImageIdentifier:   i,
			ChapterIdentifier: ch.Info.Identifier,
			VolumeIdentifier:  ch.Info.VolumeIdentifier,
		})
	}

	return result
}

type multiple []string

func (s multiple) String() string {
	if len(s) == 0 {
		return "Unknown"
	}

	return strings.Join(s, " and ")
}

func reverse(v interface{}) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(v)
		swp := reflect.Swapper(v)
		for i, j := 0, s.Len()-1; i < j; i, j = i+1, j-1 {
			swp(i, j)
		}
	default:
		panic("not a slice")
	}
}

func first(m map[string]string) string {
	for _, val := range m {
		return val
	}

	panic("empty map")
}

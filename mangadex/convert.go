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
	for _, a := range authors.Data {
		authorNames = append(authorNames, a.Attributes.Name)
	}

	artistNames := make([]string, 0)
	for _, a := range artists.Data {
		artistNames = append(artistNames, a.Attributes.Name)
	}

	return MangaInfo{
		Title:   first(b.Data.Attributes.Title),
		Authors: authorNames,
		Artists: artistNames,
		ID:      b.Data.ID,
	}
}

func convertChapters(ca []api.ChapterData, groupMap map[string]api.GroupData) ChapterList {
	sorted := make(ChapterList, 0)
	for _, info := range ca {
		lang, _ := language.Parse(info.Attributes.TranslatedLanguage)
		groups := make([]string, 0)
		for _, id := range info.Relationships.Group {
			groups = append(groups, groupMap[id].Attributes.Name)
		}

		sorted = append(sorted, Chapter{
			Info: ChapterInfo{
				Title:            info.Attributes.Title,
				Language:         lang,
				Views:            0, // FIXME
				GroupNames:       groups,
				Published:        info.Attributes.PublishAt,
				ID:               info.ID,
				Identifier:       NewWithFallback(info.Attributes.Chapter, info.Attributes.Title),
				VolumeIdentifier: NewWithFallback(info.Attributes.Volume, "Special"),
			},
			Pages: make(map[int]image.Image),
		})
	}

	reverse(sorted)
	return sorted
}

func convertCovers(coverBaseURL string, mangaID string, co []api.CoverData) PathList {
	result := make(PathList, 0)
	for _, info := range co {
		url := strings.Join([]string{coverBaseURL, mangaID, info.Attributes.FileName}, "/")
		result = append(result, Path{
			DataURL:           url,
			ImageIdentifier:   0,
			ChapterIdentifier: NewIdentifier("0"),
			VolumeIdentifier:  NewWithFallback(info.Attributes.Volume, "Special"),
		})
	}

	return result
}

func convertChapter(ch *Chapter, ah *api.AtHome) PathList {
	result := make(PathList, 0)
	for i := range ah.Chapter.Data {
		dataURL := strings.Join([]string{ah.BaseURL, "data", ah.Chapter.Hash, ah.Chapter.Data[i]}, "/")
		dataSaverURL := strings.Join([]string{ah.BaseURL, "data-saver", ah.Chapter.Hash, ah.Chapter.DataSaver[i]}, "/")
		result = append(result, Path{
			DataURL:           dataURL,
			DataSaverURL:      dataSaverURL,
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

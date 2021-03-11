package mangadex

import (
	"html"
	"path"
	"reflect"
	"time"

	"github.com/leotaku/kojirou/mangadex/api"
	"golang.org/x/text/language"
)

func convertBase(b api.BaseData) MangaInfo {
	return MangaInfo{
		Title:    b.Title,
		Authors:  b.Author,
		Artists:  b.Artist,
		IsHentai: b.IsHentai,
		ID:       b.ID,
	}
}

func convertCovers(co api.CoversData) PathList {
	result := make(PathList, 0)
	for id, url := range groupCovers(co) {
		result = append(result, PathItem{
			URL:      url,
			volumeID: NewIdentifier(id, "Special"),
		})
	}

	return result
}

func convertChapters(ca api.ChaptersData) ChapterList {
	sorted := make(ChapterList, 0)
	groups := groupGroups(ca.Groups)

	for _, info := range ca.Chapters {
		region, _ := language.ParseRegion(info.Language)
		sorted = append(sorted, ChapterInfo{
			Title:            info.Title,
			Region:           region,
			Views:            info.Views,
			Hash:             info.Hash,
			GroupNames:       unescape(getGroups(groups, info.Groups)),
			Published:        time.Unix(int64(info.Timestamp), 0),
			ID:               info.ID,
			Identifier:       NewIdentifier(info.Chapter, info.Title),
			VolumeIdentifier: NewIdentifier(info.Volume, "Special"),
		})
	}

	reverse(sorted)
	return sorted
}

func convertChapter(c api.ChapterData, chapterID Identifier, volumeID Identifier) PathList {
	result := make(PathList, 0)
	for i, filename := range c.Pages {
		url := c.Server + path.Join(c.Hash, filename)
		result = append(result, PathItem{
			URL:       url,
			imageID:   i,
			chapterID: chapterID,
			volumeID:  volumeID,
		})
	}

	return result
}

type groupsMapping = map[int]string

func groupGroups(gs []api.GroupMapping) groupsMapping {
	mapping := make(groupsMapping)
	for _, val := range gs {
		mapping[val.ID] = val.Name
	}
	return mapping
}

func getGroups(gs groupsMapping, ids []int) []string {
	result := make([]string, 0)
	for _, id := range ids {
		if name, ok := gs[id]; ok {
			result = append(result, name)
		}
	}
	return result
}

type coversMapping = map[string]string

func groupCovers(co api.CoversData) coversMapping {
	mapping := make(coversMapping)
	for _, val := range co {
		mapping[val.Volume] = val.URL
	}
	return mapping
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

func unescape(ss []string) []string {
	result := make([]string, 0)
	for _, it := range ss {
		result = append(result, html.UnescapeString(it))
	}

	return result
}

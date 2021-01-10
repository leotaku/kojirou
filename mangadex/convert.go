package mangadex

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/leotaku/manki/mangadex/api"
	"golang.org/x/text/language"
)

func convertBase(b api.BaseData) MangaInfo {
	return MangaInfo{
		Title:    b.Title,
		Authors:  b.Author,
		Artists:  b.Artist,
		IsHentai: b.IsHentai,
		Id:       b.Id,
	}
}

func convertCovers(co api.CoversData) PathList {
	result := make(PathList, 0)
	for id, url := range groupCovers(co) {
		result = append(result, PathItem{
			Url:      url,
			volumeId: NewIdentifier(id, "Special"),
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
			GroupNames:       getGroups(groups, info.Groups),
			Id:               info.Id,
			Identifier:       NewIdentifier(info.Chapter, info.Title),
			VolumeIdentifier: NewIdentifier(info.Volume, "Special"),
		})
	}

	reverse(sorted)
	return sorted
}

func convertChapter(c api.ChapterData, chapterId Identifier, volumeId Identifier) PathList {
	result := make(PathList, 0)
	for i, filename := range c.Pages {
		server := strings.TrimRight(c.Server, "/")
		url := fmt.Sprintf("%v/%v/%v", server, c.Hash, filename)
		result = append(result, PathItem{
			Url:       url,
			imageId:   i,
			chapterId: chapterId,
			volumeId:  volumeId,
		})
	}

	return result
}

type groupsMapping = map[int]string

func groupGroups(gs []api.GroupMapping) groupsMapping {
	mapping := make(groupsMapping)
	for _, val := range gs {
		mapping[val.Id] = val.Name
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
		mapping[val.Volume] = val.Url
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

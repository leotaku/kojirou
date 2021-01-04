package mangadex

import (
	"fmt"
	"sort"
	"strings"

	"github.com/leotaku/manki/mangadex/api"
	"golang.org/x/text/language"
)

func convert(b api.BaseData, ca api.ChaptersData, co api.CoversData) Manga {
	m := groupChapters(ca)
	gs := groupGroups(ca.Groups)
	cs := groupCovers(co)
	versions := remap(m, gs, cs)
	setMissing(ca, versions)

	return Manga{
		Title:       b.Title,
		Authors:     b.Author,
		Artists:     b.Artist,
		Description: b.Description,
		IsHentai:    b.IsHentai,
		Versions:    versions,
		Id:          b.Id,
	}
}

type mapping = map[string]map[string][]api.ChapterInfo
type groupsMapping = map[int]string
type coversMapping = map[string]string

func groupChapters(ca api.ChaptersData) mapping {
	m := make(mapping, 0)
	for _, ch := range ca.Chapters {
		gs := stringify(ch.Groups)
		vol := ch.Volume
		if v, ok := m[gs][vol]; ok {
			m[gs][vol] = append(v, ch)
		} else {
			if _, ok := m[gs]; ok {
				m[gs][vol] = []api.ChapterInfo{ch}
			} else {
				m[gs] = make(map[string][]api.ChapterInfo, 0)
				m[gs][vol] = []api.ChapterInfo{ch}
			}
		}
	}

	return m
}

func groupGroups(gs []api.GroupMapping) groupsMapping {
	mapping := make(groupsMapping, 0)
	for _, val := range gs {
		mapping[val.Id] = val.Name
	}
	return mapping
}

func groupCovers(co api.CoversData) coversMapping {
	mapping := make(coversMapping, 0)
	for _, val := range co {
		mapping[val.Volume] = val.Url
	}
	return mapping
}

func remap(m mapping, gs groupsMapping, cs coversMapping) []Version {
	versions := make([]Version, 0)
	lastId := api.ChapterInfo{}

	for _, vols := range m {
		volumes := make([]Volume, 0)
		for _, chaps := range vols {
			chapters := make([]Chapter, 0)
			for _, chap := range chaps {
				it := convertChapter(chap)
				chapters = append(chapters, it)
				lastId = chap
			}
			reverse(chapters)
			it := convertVolume(lastId, cs, chapters)
			volumes = append(volumes, it)
		}
		resort(volumes)
		it := convertVersion(lastId, gs, volumes)
		versions = append(versions, it)
	}
	resort(versions)

	return versions
}

func convertVersion(a api.ChapterInfo, gs groupsMapping, it []Volume) Version {
	region, _ := language.ParseRegion(a.Language)
	return Version{
		GroupNames: getGroups(gs, a.Groups),
		Volumes:    it,
		Missing:    nil,
		Region:     region,
	}
}

func convertVolume(a api.ChapterInfo, cs coversMapping, it []Chapter) Volume {
	coverUrl := ""
	if v, ok := cs[a.Volume]; ok {
		coverUrl = v
	}

	return Volume{
		Number:     GuessIdentifier(a.Volume),
		CoverImage: Image{Url: coverUrl},
		Chapters:   it,
	}
}

func convertChapter(it api.ChapterInfo) Chapter {
	return Chapter{
		Title:  it.Title,
		Number: GuessIdentifier(it.Chapter),
		Views:  it.Views,
		Hash:   it.Hash,
		Id:     it.Id,
	}
}

func setMissing(ca api.ChaptersData, vs []Version) {
	for I, version := range vs {
		has := make(map[Identifier]struct{}, 0)
		for _, volume := range version.Volumes {
			for _, chap := range volume.Chapters {
				has[chap.Number] = struct{}{}
			}
		}

		for _, chap := range ca.Chapters {
			num := GuessIdentifier(chap.Chapter)
			if _, ok := has[num]; !ok {
				version.Missing = append(version.Missing, num)
				has[num] = struct{}{}
			}
		}
		resort(version.Missing)
		vs[I].Missing = version.Missing
	}
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

func stringify(ids []int) string {
	b := new(strings.Builder)
	for _, id := range ids {
		fmt.Fprintf(b, "%v", id)
	}
	return b.String()
}

func reverse(v interface{}) {
	switch s := v.(type) {
	case []Chapter:
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
		}
	case []Volume:
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
		}
	default:
		panic("reverse: Unsupported type")
	}
}

func resort(v interface{}) {
	switch s := v.(type) {
	case []Identifier:
		sort.Slice(s, func(i, j int) bool {
			return s[i].Less(s[j])
		})
	case []Chapter:
		sort.Slice(s, func(i, j int) bool {
			return s[i].Number.Less(s[j].Number)
		})
	case []Volume:
		sort.Slice(s, func(i, j int) bool {
			return s[i].Number.Less(s[j].Number)
		})
	case []Version:
		sort.Slice(s, func(i, j int) bool {
			gi := strings.Join(s[i].GroupNames, ", ")
			gj := strings.Join(s[j].GroupNames, ", ")
			return gi < gj
		})
	default:
		panic("resort: Unsupported type")
	}
}

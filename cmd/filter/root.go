package filter

import (
	"fmt"
	"reflect"
	"time"

	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

type Filter = func(md.ChapterList) (md.ChapterList, error)

func FilterByLanguage(cl md.ChapterList, lang language.Tag) md.ChapterList {
	return cl.FilterBy(func(c md.ChapterInfo) bool {
		return c.Language == lang
	})
}

func FilterByRegex(cl md.ChapterList, field string, pattern string) md.ChapterList {
	return cl.FilterBy(func(ci md.ChapterInfo) bool {
		v := reflect.ValueOf(ci).FieldByName(field).Interface()
		return MatchPattern(pattern, fmt.Sprint(v))
	})
}

func FilterByIdentifier(cl md.ChapterList, field string, values []Range) md.ChapterList {
	return cl.FilterBy(func(ci md.ChapterInfo) bool {
		v := reflect.ValueOf(ci).FieldByName(field).Interface()
		switch f := v.(type) {
		case md.Identifier:
			for _, r := range values {
				ok := r.Contains(f)
				if ok {
					return true
				}
			}
		default:
			panic("field is not identifier")
		}

		return false
	})
}

func SortByNewest(cl md.ChapterList) md.ChapterList {
	return cl.SortBy(func(ci1, ci2 md.ChapterInfo) bool {
		return ci1.Published.After(ci2.Published)
	})
}

func SortByNewestGroup(cl md.ChapterList) md.ChapterList {
	groupRanking := make(map[string]time.Time)
	for _, c := range cl {
		if val, ok := groupRanking[gid(c.Info)]; !ok || c.Info.Published.Before(val) {
			groupRanking[gid(c.Info)] = c.Info.Published
		}
	}

	return cl.SortBy(func(ci1, ci2 md.ChapterInfo) bool {
		return groupRanking[gid(ci1)].After(groupRanking[gid(ci2)])
	})
}

func SortByViews(cl md.ChapterList) md.ChapterList {
	return cl.SortBy(func(ci1, ci2 md.ChapterInfo) bool {
		return ci1.Views > ci2.Views
	})
}

func SortByGroupViews(cl md.ChapterList) md.ChapterList {
	groupRanking := make(map[string]int)
	for _, ci := range cl {
		groupRanking[gid(ci.Info)] += ci.Info.Views
	}

	return cl.SortBy(func(ci1, ci2 md.ChapterInfo) bool {
		return groupRanking[gid(ci1)] > groupRanking[gid(ci2)]
	})
}

func SortByMost(cl md.ChapterList) md.ChapterList {
	groupRanking := make(map[string]int)
	for _, ci := range cl {
		groupRanking[gid(ci.Info)] += 1
	}

	return cl.SortBy(func(ci1, ci2 md.ChapterInfo) bool {
		return groupRanking[gid(ci1)] > groupRanking[gid(ci2)]
	})
}

func RemoveDuplicates(cl md.ChapterList) md.ChapterList {
	return cl.CollapseBy(func(c md.ChapterInfo) interface{} {
		return struct {
			chapter md.Identifier
			volume  md.Identifier
		}{
			c.Identifier,
			c.VolumeIdentifier,
		}
	})
}

func gid(ci md.ChapterInfo) string {
	return ci.GroupNames.String()
}

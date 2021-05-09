package cmd

import (
	"fmt"
	"reflect"
	"time"

	"github.com/leotaku/kojirou/cmd/util"
	"github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
)

type Filter = func(mangadex.ChapterList) (mangadex.ChapterList, error)

func filterLanguage(cl mangadex.ChapterList, lang language.Tag) mangadex.ChapterList {
	return cl.FilterBy(func(c mangadex.ChapterInfo) bool {
		return c.Language == lang
	})
}

func filterRegexField(cl mangadex.ChapterList, field string, pattern string) mangadex.ChapterList {
	return cl.FilterBy(func(ci mangadex.ChapterInfo) bool {
		v := reflect.ValueOf(ci).FieldByName(field).Interface()
		return util.Match(pattern, fmt.Sprint(v))
	})
}

func filterIdentifierField(cl mangadex.ChapterList, field string, values []util.Range) mangadex.ChapterList {
	return cl.FilterBy(func(ci mangadex.ChapterInfo) bool {
		v := reflect.ValueOf(ci).FieldByName(field).Interface()
		switch f := v.(type) {
		case mangadex.Identifier:
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

func rankNewest(cl mangadex.ChapterList) mangadex.ChapterList {
	return cl.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return ci1.Published.After(ci2.Published)
	})
}

func rankTotalNewest(cl mangadex.ChapterList) mangadex.ChapterList {
	groupRanking := make(map[string]time.Time)
	for _, ci := range cl {
		if val, ok := groupRanking[gid(ci)]; !ok || ci.Published.Before(val) {
			groupRanking[gid(ci)] = ci.Published
		}
	}

	return cl.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return groupRanking[gid(ci1)].After(groupRanking[gid(ci2)])
	})
}

func rankViews(cl mangadex.ChapterList) mangadex.ChapterList {
	return cl.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return ci1.Views > ci2.Views
	})
}

func rankTotalViews(cl mangadex.ChapterList) mangadex.ChapterList {
	groupRanking := make(map[string]int)
	for _, ci := range cl {
		groupRanking[gid(ci)] += ci.Views
	}

	return cl.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return groupRanking[gid(ci1)] > groupRanking[gid(ci2)]
	})
}

func rankMost(cl mangadex.ChapterList) mangadex.ChapterList {
	groupRanking := make(map[string]int)
	for _, ci := range cl {
		groupRanking[gid(ci)] += 1
	}

	return cl.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return groupRanking[gid(ci1)] > groupRanking[gid(ci2)]
	})
}

func doRank(cl mangadex.ChapterList) mangadex.ChapterList {
	cl = cl.CollapseBy(func(c mangadex.ChapterInfo) interface{} {
		return c.Identifier
	})
	return cl.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return ci1.Identifier.Less(ci2.Identifier)
	})
}

func gid(ci mangadex.ChapterInfo) string {
	return ci.GroupNames.String()
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/leotaku/manki/cmd/util"
	"github.com/leotaku/manki/mangadex"
	"golang.org/x/text/language"
)

func filter(m mangadex.ChapterList, lang language.Tag) (*mangadex.ChapterList, error) {
	// Filter group by language
	m = m.FilterBy(func(c mangadex.ChapterInfo) bool {
		return util.MatchRegion(c.Region) == lang
	})

	// Rank groups by total views
	gid := func(ci mangadex.ChapterInfo) string {
		return strings.Join(ci.GroupNames, "")
	}
	groupRanking := make(map[string]int)
	for _, ci := range m {
		groupRanking[gid(ci)] += ci.Views
	}

	// Sort, collapse and sort again
	m = m.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return groupRanking[gid(ci1)] > groupRanking[gid(ci2)]
	})
	m = m.CollapseBy(func(c mangadex.ChapterInfo) interface{} {
		return c.Identifier
	})
	m = m.SortBy(func(ci1, ci2 mangadex.ChapterInfo) bool {
		return ci1.Identifier.Less(ci2.Identifier)
	})

	if len(m) > 0 {
		return &m, nil
	} else {
		return nil, fmt.Errorf("No matching scantlations found")
	}
}

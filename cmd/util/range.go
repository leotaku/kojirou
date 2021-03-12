package util

import (
	"strings"

	"github.com/leotaku/kojirou/mangadex"
)

type Range struct {
	start mangadex.Identifier
	end   *mangadex.Identifier
}

func ParseRanges(s string) []Range {
	result := make([]Range, 0)
	for _, it := range strings.Split(s, ",") {
		if se := strings.Split(it, ".."); len(se) == 2 {
			start := mangadex.NewIdentifier(se[0], "")
			end := mangadex.NewIdentifier(se[1], "")
			result = append(result, Range{
				start: start,
				end:   &end,
			})
		} else {
			result = append(result, Range{
				start: mangadex.NewIdentifier(it, it),
			})
		}
	}

	return result
}

func (r *Range) Contains(id mangadex.Identifier) bool {
	if r.end != nil {
		return r.start.LessOrEqual(id) && id.LessOrEqual(*r.end)
	} else {
		return r.start.Equal(id)
	}
}

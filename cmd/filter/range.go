package filter

import (
	"strings"

	md "github.com/leotaku/kojirou/mangadex"
)

type Range struct {
	start md.Identifier
	end   *md.Identifier
}

func ParseRanges(s string) []Range {
	result := make([]Range, 0)
	for _, it := range strings.Split(s, ",") {
		if se := strings.Split(it, ".."); len(se) == 2 {
			start := md.NewIdentifier(se[0])
			end := md.NewIdentifier(se[1])
			result = append(result, Range{
				start: start,
				end:   &end,
			})
		} else {
			result = append(result, Range{
				start: md.NewIdentifier(it),
			})
		}
	}

	return result
}

func (r *Range) Contains(id md.Identifier) bool {
	if r.end != nil {
		return r.start.LessOrEqual(id) && id.LessOrEqual(*r.end)
	} else {
		return r.start.Equal(id)
	}
}

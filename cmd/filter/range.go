package filter

import (
	"strings"

	md "github.com/leotaku/kojirou/mangadex"
)

type Ranges struct {
	ranges  []singleRange
	negated bool
}

func ParseRanges(s string) Ranges {
	if strings.HasPrefix(s, "!") {
		return Ranges{
			ranges:  parseRangeList(s[1:]),
			negated: true,
		}
	} else {
		return Ranges{
			ranges:  parseRangeList(s),
			negated: false,
		}
	}
}

func (rs *Ranges) Contains(id md.Identifier) bool {
	for _, r := range rs.ranges {
		ok := r.contains(id)
		if ok {
			return !rs.negated
		}
	}

	return rs.negated
}

type singleRange struct {
	start md.Identifier
	end   *md.Identifier
}

func parseRangeList(s string) []singleRange {
	ranges := make([]singleRange, 0)
	for _, it := range strings.Split(s, ",") {
		if se := strings.Split(it, ".."); len(se) == 2 {
			start := md.NewIdentifier(se[0])
			end := md.NewIdentifier(se[1])
			ranges = append(ranges, singleRange{
				start: start,
				end:   &end,
			})
		} else {
			ranges = append(ranges, singleRange{
				start: md.NewIdentifier(it),
				end:   nil,
			})
		}
	}

	return ranges
}

func (r *singleRange) contains(id md.Identifier) bool {
	if r.end != nil {
		return r.start.LessOrEqual(id) && id.LessOrEqual(*r.end)
	} else {
		return r.start.Equal(id)
	}
}

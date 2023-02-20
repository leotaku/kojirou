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
		if ok := r.contains(id); ok {
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
	for _, rangeExpr := range strings.Split(s, ",") {
		if startAndEnd := strings.Split(rangeExpr, ".."); len(startAndEnd) == 2 {
			start := md.NewIdentifier(startAndEnd[0])
			end := md.NewIdentifier(startAndEnd[1])
			ranges = append(ranges, singleRange{
				start: start,
				end:   &end,
			})
		} else {
			ranges = append(ranges, singleRange{
				start: md.NewIdentifier(rangeExpr),
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

package filter

import (
	"regexp"
	"strings"
)

func MatchPattern(pattern, s string) bool {
	negate := strings.HasPrefix(pattern, "!")
	if negate {
		matched, _ := regexp.MatchString(pattern[1:], s)
		return !matched
	} else {
		matched, _ := regexp.MatchString(pattern, s)
		return matched
	}
}

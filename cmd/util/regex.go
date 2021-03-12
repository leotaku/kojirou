package util

import (
	"regexp"
	"strings"
)

func Match(pattern, s string) bool {
	negate := strings.HasPrefix(pattern, "!")
	if negate {
		matched, _ := regexp.MatchString(pattern[1:], s)
		return !matched
	} else {
		matched, _ := regexp.MatchString(pattern, s)
		return matched
	}
}

package filter

import (
	"github.com/leotaku/mobi"
	"golang.org/x/text/language"
)

var matcher = language.NewMatcher(mobi.SupportedLocales)

func MatchLang(s string) language.Tag {
	lang := language.Make(s)
	_, i, _ := matcher.Match(lang)
	return mobi.SupportedLocales[i]
}

func MatchRegion(region language.Region) language.Tag {
	lang, _ := language.Compose(region)
	_, i, _ := matcher.Match(lang)
	return mobi.SupportedLocales[i]
}

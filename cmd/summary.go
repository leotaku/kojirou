package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	md "github.com/leotaku/kojirou/mangadex"
)

var groupColors = []*color.Color{
	color.New(color.FgRed),
	color.New(color.FgBlue),
	color.New(color.FgMagenta),
	color.New(color.FgCyan),
	color.New(color.FgGreen),
	color.New(color.FgYellow),
	color.New(color.ReverseVideo, color.FgRed),
	color.New(color.ReverseVideo, color.FgBlue),
	color.New(color.ReverseVideo, color.FgMagenta),
	color.New(color.ReverseVideo, color.FgCyan),
	color.New(color.ReverseVideo, color.FgGreen),
	color.New(color.ReverseVideo, color.FgYellow),
}

func printMangaSummary(manga *md.Manga) {
	sorted := manga.Chapters().SortBy(func(a md.ChapterInfo, b md.ChapterInfo) bool {
		return a.Identifier.Less(b.Identifier)
	})
	groups, numbers := formatChapterMapping(sorted)
	discontinuities := formatDiscontinuities(sorted)

	printValue("Title", manga.Info.Title)
	printValue("Author", manga.Info.Authors)
	if len(numbers) > 0 {
		printValue("Groups", strings.Join(groups, ", "))
		printValue("Chapters", strings.Join(numbers, ", "))
	}
	if len(discontinuities) > 0 {
		printValue("Discontinuities", strings.Join(discontinuities, ", "))
	}
}

func formatChapterMapping(chapters md.ChapterList) (groups, numbers []string) {
	colorIndices := make(map[string]int)
	for _, chapter := range chapters {
		group := chapter.Info.GroupNames.String()
		if index, ok := colorIndices[group]; !ok {
			index = len(colorIndices)
			colorIndices[group] = index
			color := groupColors[index%len(groupColors)]

			// Small code duplication for clearer code
			groups = append(groups, color.Sprint(group))
			numbers = append(numbers, color.Sprint(chapter.Info.Identifier))
		} else {
			color := groupColors[index%len(groupColors)]

			// Small code duplication for clearer code
			numbers = append(numbers, color.Sprint(chapter.Info.Identifier))
		}
	}

	return groups, numbers
}

func formatDiscontinuities(chapters md.ChapterList) (discontinuities []string) {
	last := md.NewIdentifier("0")
	for i := range chapters {
		this := chapters[i].Info.Identifier
		if this.IsSpecial() || last.IsSpecial() {
			continue
		}

		if !last.Equal(this) && !last.IsNext(this) {
			discontinuities = append(discontinuities, fmt.Sprintf("%v..%v", last, this))
		}
		last = this
	}

	return discontinuities
}

func printValue(name, value interface{}) {
	underlined := color.New(color.Underline)
	fmt.Printf("%v: %v\n", underlined.Sprint(name), value)
}

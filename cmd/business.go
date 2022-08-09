package cmd

import (
	"fmt"

	"github.com/leotaku/kojirou/cmd/crop"
	"github.com/leotaku/kojirou/cmd/filter"
	"github.com/leotaku/kojirou/cmd/formats"
	"github.com/leotaku/kojirou/cmd/formats/download"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
)

func run() error {
	manga, err := download.MangadexSkeleton(identifierArg)
	if err != nil {
		return fmt.Errorf("skeleton: %w", err)
	}
	chapters, err := download.MangadexChapters(identifierArg)
	if err != nil {
		return fmt.Errorf("chapters: %w", err)
	}
	chapters, err = filterFromFlags(chapters)
	if err != nil {
		return fmt.Errorf("filter: %w", err)
	}
	*manga = manga.WithChapters(chapters)

	formats.PrintSummary(manga)
	if dryRunArg {
		return nil
	}

	r := formats.VanishingProgress("Covers")
	covers, err := download.MangadexCovers(manga, r)
	r.Done()
	if err != nil {
		return fmt.Errorf("covers: %w", err)
	}
	*manga = manga.WithCovers(covers)

	dir := kindle.NewNormalizedDirectory(outArg, manga.Info.Title, kindleFolderModeArg)
	for _, volume := range manga.Sorted() {
		r := formats.TitledProgress(fmt.Sprintf("Volume: %v", volume.Info.Identifier))
		if dir.Has(volume.Info.Identifier) && !forceArg {
			r.Cancel("Skipped")
			continue
		}
		pages, err := download.MangadexPages(volume.Sorted(), r)
		r.Done()
		if err != nil {
			return fmt.Errorf("pages: %w", err)
		}
		if autocropArg {
			r := formats.VanishingProgress("Cropping")
			autoCrop(pages, r)
			r.Done()
		}
		part := manga.WithChapters(volume.Sorted()).WithPages(pages)
		r = formats.VanishingProgress("Writing...")
		if err := dir.Write(part, r); err != nil {
			r.Cancel("Failed")
			return fmt.Errorf("write: %w", err)
		}
		r.Done()
	}

	return nil
}

func autoCrop(pages md.ImageList, r formats.Reporter) error {
	for i, page := range pages {
		if cropped, err := crop.Crop(pages[i].Image, crop.Limited(pages[i].Image, 0.1)); err != nil {
			return fmt.Errorf("chapter %v: page %v: %w", page.ChapterIdentifier, page.ImageIdentifier, err)
		} else {
			pages[i].Image = cropped
			r.Add(1)
		}
	}

	return nil
}

func filterFromFlags(cl md.ChapterList) (md.ChapterList, error) {
	if languageArg != "" {
		lang := filter.MatchLang(languageArg)
		cl = filter.FilterByLanguage(cl, lang)
	}
	if groupsFilter != "" {
		cl = filter.FilterByRegex(cl, "GroupNames", groupsFilter)
	}
	if volumesFilter != "" {
		ranges := filter.ParseRanges(volumesFilter)
		cl = filter.FilterByIdentifier(cl, "VolumeIdentifier", ranges)
	}
	if chaptersFilter != "" {
		ranges := filter.ParseRanges(chaptersFilter)
		cl = filter.FilterByIdentifier(cl, "Identifier", ranges)
	}

	switch rankArg {
	case "newest":
		cl = filter.SortByNewest(cl)
	case "newest-total":
		cl = filter.SortByNewestGroup(cl)
	case "views":
		cl = filter.SortByViews(cl)
	case "views-total":
		cl = filter.SortByGroupViews(cl)
	case "most":
		cl = filter.SortByMost(cl)
	default:
		return nil, fmt.Errorf(`not a valid rankinging algorithm: "%v"`, rankArg)
	}

	return filter.RemoveDuplicates(cl), nil
}

package cmd

import (
	"fmt"

	"github.com/leotaku/kojirou/cmd/crop"
	"github.com/leotaku/kojirou/cmd/filter"
	"github.com/leotaku/kojirou/cmd/formats"
	"github.com/leotaku/kojirou/cmd/formats/disk"
	"github.com/leotaku/kojirou/cmd/formats/download"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/text/language"
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

	if diskArg != "" {
		p := formats.VanishingProgress("Disk...")
		diskChapters, err := disk.LoadChapters(diskArg, language.Make(languageArg), p)
		if err != nil {
			p.Cancel("Error")
			return fmt.Errorf("disk: %w", err)
		}
		p.Done()
		chapters = append(chapters, diskChapters...)
	}

	chapters, err = sortFromFlags(chapters)
	if err != nil {
		return fmt.Errorf("filter: %w", err)
	}

	// Ensure chapters from disk are preferred
	if diskArg != "" {
		chapters.SortBy(func(a md.ChapterInfo, b md.ChapterInfo) bool {
			return a.GroupNames.String() == "Filesystem" && b.GroupNames.String() != "Filesystem"
		})
	}

	chapters = filter.RemoveDuplicates(chapters)
	*manga = manga.WithChapters(chapters)
	formats.PrintSummary(manga)
	if dryRunArg {
		return nil
	}

	p := formats.VanishingProgress("Covers")
	covers, err := download.MangadexCovers(manga, p)
	if err != nil {
		p.Cancel("Error")
		return fmt.Errorf("covers: %w", err)
	}
	p.Done()

	// Covers from disk should automatically be preferred, because
	// they appear later in the list and thus should override the
	// earlier downloaded covers.
	if diskArg != "" {
		p := formats.VanishingProgress("Disk...")
		diskCovers, err := disk.LoadCovers(diskArg, p)
		if err != nil {
			p.Cancel("Error")
			return fmt.Errorf("disk: %w", err)
		}
		p.Done()
		covers = append(covers, diskCovers...)
	}
	*manga = manga.WithCovers(covers)

	dir := kindle.NewNormalizedDirectory(outArg, manga.Info.Title, kindleFolderModeArg)
	for _, volume := range manga.Sorted() {
		p := formats.TitledProgress(fmt.Sprintf("Volume: %v", volume.Info.Identifier))
		if dir.Has(volume.Info.Identifier) && !forceArg {
			p.Cancel("Skipped")
			continue
		}
		mangadexPages, err := download.MangadexPages(volume.Sorted().FilterBy(func(ci md.ChapterInfo) bool {
			return ci.GroupNames.String() != "Filesystem"
		}), p)
		if err != nil {
			p.Cancel("Error")
			return fmt.Errorf("pages: %w", err)
		}
		diskPages, err := disk.LoadPages(volume.Sorted().FilterBy(func(ci md.ChapterInfo) bool {
			return ci.GroupNames.String() == "Filesystem"
		}), p)
		if err != nil {
			p.Cancel("Error")
			return fmt.Errorf("pages: %w", err)
		}
		p.Done()
		pages := append(mangadexPages, diskPages...)
		if autocropArg {
			r := formats.VanishingProgress("Cropping..")
			if err := autoCrop(pages, r); err != nil {
				return fmt.Errorf("autocrop: %w", err)
			}
			r.Done()
		}
		part := manga.WithChapters(volume.Sorted()).WithPages(pages)
		p = formats.VanishingProgress("Writing...")
		if err := dir.Write(part, p); err != nil {
			p.Cancel("Failed")
			return fmt.Errorf("write: %w", err)
		}
		p.Done()
	}

	return nil
}

func autoCrop(pages md.ImageList, p formats.Progress) error {
	p.Increase(len(pages))
	for i, page := range pages {
		if cropped, err := crop.Crop(pages[i].Image, crop.Limited(pages[i].Image, 0.1)); err != nil {
			return fmt.Errorf("chapter %v: page %v: %w", page.ChapterIdentifier, page.ImageIdentifier, err)
		} else {
			pages[i].Image = cropped
			p.Add(1)
		}
	}

	return nil
}

func sortFromFlags(cl md.ChapterList) (md.ChapterList, error) {
	if languageArg != "" {
		lang := language.Make(languageArg)
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

	return cl, nil
}

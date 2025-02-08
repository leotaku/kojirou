package cmd

import (
	"fmt"

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

	chapters, err := getChapters(*manga)
	if err != nil {
		return fmt.Errorf("chapters: %w", err)
	}
	*manga = manga.WithChapters(chapters)

	formats.PrintSummary(manga)
	if dryRunArg {
		return nil
	}

	covers, err := getCovers(manga)
	if err != nil {
		return fmt.Errorf("covers: %w", err)
	}
	*manga = manga.WithCovers(covers)

	dir := kindle.NewNormalizedDirectory(outArg, manga.Info.Title, kindleFolderModeArg)
	for _, volume := range manga.Sorted() {
		if err := handleVolume(*manga, volume, dir); err != nil {
			return fmt.Errorf("volume %v: %w", volume.Info.Identifier, err)
		}
	}

	return nil
}

func handleVolume(skeleton md.Manga, volume md.Volume, dir kindle.NormalizedDirectory) error {
	p := formats.TitledProgress(fmt.Sprintf("Volume: %v", volume.Info.Identifier))
	if dir.Has(volume.Info.Identifier) && !forceArg {
		p.Cancel("Skipped")
		return nil
	}

	pages, err := getPages(volume, skeleton.Info.Title, p)
	if err != nil {
		return fmt.Errorf("pages: %w", err)
	}

	mangaForVolume := skeleton.WithChapters(volume.Sorted()).WithPages(pages)
	mobi := kindle.GenerateMOBI(
		mangaForVolume,
		kindle.WidepagePolicy(widepageArg),
		autocropArg,
		leftToRightArg,
	)
	mobi.RightToLeft = !leftToRightArg
	mobi.Title = fmt.Sprintf("%v: %v",
		skeleton.Info.Title,
		volume.Info.Identifier.StringFilled(fillVolumeNumberArg, 0, false),
	)

	p = formats.VanishingProgress("Writing...")
	if err := dir.Write(volume.Info.Identifier, mobi, p); err != nil {
		p.Cancel("Error")
		return fmt.Errorf("write: %w", err)
	}
	p.Done()

	return nil
}

func getChapters(manga md.Manga) (md.ChapterList, error) {
	chapters, err := download.MangadexChapters(manga.Info.ID)
	if err != nil {
		return nil, fmt.Errorf("mangadex: %w", err)
	}

	if diskArg != "" {
		p := formats.VanishingProgress("Disk...")
		diskChapters, err := disk.LoadChapters(diskArg, language.Make(languageArg), p)
		if err != nil {
			p.Cancel("Error")
			return nil, fmt.Errorf("disk: %w", err)
		}
		p.Done()
		chapters = append(chapters, diskChapters...)
	}

	chapters, err = filterAndSortFromFlags(chapters)
	if err != nil {
		return nil, fmt.Errorf("filter: %w", err)
	}

	// Ensure chapters from disk are preferred
	if diskArg != "" {
		chapters = chapters.SortBy(func(a md.ChapterInfo, b md.ChapterInfo) bool {
			return a.GroupNames.String() == "Filesystem" && b.GroupNames.String() != "Filesystem"
		})
	}

	return filter.RemoveDuplicates(chapters), nil
}

func getCovers(manga *md.Manga) (md.ImageList, error) {
	p := formats.VanishingProgress("Covers")
	covers, err := download.MangadexCovers(manga, saveRawArg, fillVolumeNumberArg, p)
	if err != nil {
		p.Cancel("Error")
		return nil, fmt.Errorf("mangadex: %w", err)
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
			return nil, fmt.Errorf("disk: %w", err)
		}
		p.Done()
		covers = append(covers, diskCovers...)
	}

	return covers, nil
}

func getPages(volume md.Volume, mangaTitle string, p formats.CliProgress) (md.ImageList, error) {
	mangadexPages, err := download.MangadexPages(volume.Sorted().FilterBy(func(ci md.ChapterInfo) bool {
		return ci.GroupNames.String() != "Filesystem"
	}), download.DataSaverPolicy(dataSaverArg), saveRawArg, fillVolumeNumberArg, mangaTitle, p)
	if err != nil {
		p.Cancel("Error")
		return nil, fmt.Errorf("mangadex: %w", err)
	}
	diskPages, err := disk.LoadPages(volume.Sorted().FilterBy(func(ci md.ChapterInfo) bool {
		return ci.GroupNames.String() == "Filesystem"
	}), p)
	if err != nil {
		p.Cancel("Error")
		return nil, fmt.Errorf("disk: %w", err)
	}
	p.Done()

	return append(mangadexPages, diskPages...), nil
}

func filterAndSortFromFlags(cl md.ChapterList) (md.ChapterList, error) {
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
		return nil, fmt.Errorf(`not a valid ranking algorithm: "%v"`, rankArg)
	}

	return cl, nil
}

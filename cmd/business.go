package cmd

import (
	"fmt"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
)

const (
	progressTemplate = `{{ string . "prefix" | printf "%-10v" }} {{ bar . "|" "█" "▌" " " "|" }} {{ counters . | printf "%-15v" }} {{ "|" }}`
)

func runBusinessLogic(mangaID string) error {
	// Create API client
	retry := retryablehttp.NewClient()
	retry.Logger = nil
	http := retry.StandardClient()
	client := md.NewClient().WithHTTPClient(http)

	manga, err := downloadManga(client, mangaID)
	if err != nil {
		return err
	}
	printMangaSummary(*manga)

	if dryRunArg {
		return nil
	}

	for _, volume := range manga.Sorted() {
		bar := pb.New(0).SetTemplate(progressTemplate)
		bar.Set("prefix", fmt.Sprintf("Volume: %v", volume.Info.Identifier))
		dl := formats.NewMangadexDownloader(client, http, progress(bar))

		bar.Start()
		chapters := volume.Sorted()
		pages, err := formats.MangadexPages(dl, chapters)
		if err != nil {
			return err
		}
		bar.Finish()

		bar = pb.New(0).SetTemplate(progressTemplate)
		bar.Set("prefix", "Writing...")
		bar.Set(pb.CleanOnFinish, true)

		bar.Start()
		f, err := os.Create(volume.Info.Identifier.StringFilled(4, 2, false) + ".azw3")
		if err != nil {
			return fmt.Errorf("create volume %s: %w", volume.Info.Identifier, err)
		}
		manga := manga.WithChapters(chapters).WithPages(pages)
		if err := formats.WriteMOBI(manga).Realize().Write(f); err != nil {
			return fmt.Errorf("write volume %s: %w", volume.Info.Identifier, err)
		}

		bar.Finish()
	}

	return nil
}

func progress(bar *pb.ProgressBar) formats.Reporter {
	return func(n int) {
		if n > 0 {
			bar.AddTotal(int64(n))
		} else {
			bar.Add(-n)
		}
	}
}

func downloadManga(client *md.Client, mangaID string) (*md.Manga, error) {
	// Fetch manga and chapter data
	manga, err := client.FetchManga(mangaID)
	if err != nil {
		return nil, fmt.Errorf("download manga: %w", err)
	}
	chapters, err := client.FetchChapters(mangaID)
	if err != nil {
		return nil, fmt.Errorf("download chapters: %w", err)
	}

	// Filter chapters
	chapters, err = filterFromFlags(chapters)
	if err != nil {
		return nil, fmt.Errorf("filter: %w", err)
	}

	result := manga.WithChapters(chapters)
	return &result, nil
}

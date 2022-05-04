package cmd

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/flesnuk/boundingbox"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
)

const (
	progressTemplate = `` +
		`{{ string . "prefix" | printf "%-10v" }}` +
		`{{ bar . "|" "█" "▌" " " "|" }}` + `{{ " " }}` +
		`{{ if string . "message" }}` +
		`{{   string . "message" | printf "%-15v" }}` +
		`{{ else }}` +
		`{{   counters . | printf "%-15v" }}` +
		`{{ end }}` + `{{ " |" }}`
)

func runBusinessLogic(mangaID string) error {
	retry := retryablehttp.NewClient()
	retry.Logger = nil
	http := retry.StandardClient()
	client := md.NewClient().WithHTTPClient(http)

	manga, err := businessDownloadManga(client, mangaID)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	printMangaSummary(manga)
	if len(manga.Volumes) == 0 {
		return fmt.Errorf("no matching chapters")
	} else if dryRunArg {
		return nil
	}

	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", "Covers")
	bar.Set(pb.CleanOnFinish, true)
	bar.Start()
	dl := formats.NewMangadexDownloader(client, http, progress(bar))
	covers, err := formats.MangadexCovers(dl, manga)
	if err != nil {
		return err
	}
	*manga = manga.WithCovers(covers)
	bar.Finish()

	bookDirectory := pathnameFromTitle(manga.Info.Title)
	thumbnailDirectory := "/dev/null"
	switch {
	case kindleFolderModeArg && outArg != "":
		bookDirectory = path.Join(outArg, "documents", bookDirectory)
		thumbnailDirectory = path.Join(outArg, "system", "thumbnails")
	case kindleFolderModeArg:
		bookDirectory = path.Join("kindle", "documents", bookDirectory)
		thumbnailDirectory = path.Join("kindle", "system", "thumbnails")
	case outArg != "":
		bookDirectory = outArg
	}

	for _, volume := range manga.Sorted() {
		bar := pb.New(0).SetTemplate(progressTemplate)
		bar.Set("prefix", fmt.Sprintf("Volume: %v", volume.Info.Identifier))
		bar.Start()

		bookFilename := path.Join(bookDirectory, volume.Info.Identifier.StringFilled(4, 2, false)+".azw3")
		switch _, err := os.Stat(bookFilename); {
		case !(os.IsNotExist(err) || forceArg):
			bar.SetTotal(1).SetCurrent(1)
			bar.Set("message", "Skipped")
			bar.Finish()
		default:
			dl := formats.NewMangadexDownloader(client, http, progress(bar))
			chapters := volume.Sorted()
			pages, err := formats.MangadexPages(dl, chapters)
			bar.Finish()

			if err == nil && autocropFlag {
				autocropImages(pages)
			}

			if err != nil {
				return fmt.Errorf("download: %w", err)
			} else if err := businessWriteBook(
				manga.WithChapters(chapters).WithPages(pages),
				bookFilename,
				thumbnailDirectory,
			); err != nil {
				return fmt.Errorf("write: %w", err)
			}
		}
	}

	return nil
}

func businessWriteBook(manga md.Manga, bookFilename string, thumbnailDirectory string) error {
	mobi := formats.WriteMOBI(manga)
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", "Writing...")
	bar.Set(pb.CleanOnFinish, true)
	bar.Start()
	defer bar.Finish()

	if err := os.MkdirAll(path.Dir(bookFilename), os.ModePerm); err != nil {
		return err
	} else if f, err := os.Create(bookFilename); err != nil {
		return err
	} else if err := mobi.Realize().Write(bar.NewProxyWriter(f)); err != nil {
		return err
	}

	if thumbnailDirectory == "/dev/null" || mobi.CoverImage == nil {
		return nil
	} else if err := os.MkdirAll(thumbnailDirectory, os.ModePerm); err != nil {
		return err
	} else if t, err := os.Create(path.Join(thumbnailDirectory, mobi.GetThumbFilename())); err != nil {
		return err
	} else if err := jpeg.Encode(t, mobi.CoverImage, nil); err != nil {
		return err
	}

	return nil
}

func businessDownloadManga(client *md.Client, mangaID string) (*md.Manga, error) {
	manga, err := client.FetchManga(mangaID)
	if err != nil {
		return nil, fmt.Errorf("manga: %w", err)
	}
	chapters, err := client.FetchChapters(mangaID)
	if err != nil {
		return nil, fmt.Errorf("chapters: %w", err)
	}

	chapters, err = filterFromFlags(chapters)
	if err != nil {
		return nil, fmt.Errorf("filter: %w", err)
	}

	result := manga.WithChapters(chapters)
	return &result, nil
}

func pathnameFromTitle(filename string) string {
	switch runtime.GOOS {
	case "windows":
		filename = strings.ReplaceAll(filename, "\"", "＂")
		filename = strings.ReplaceAll(filename, "\\", "＼")
		filename = strings.ReplaceAll(filename, "<", "＜")
		filename = strings.ReplaceAll(filename, ">", "＞")
		filename = strings.ReplaceAll(filename, ":", "：")
		filename = strings.ReplaceAll(filename, "|", "｜")
		filename = strings.ReplaceAll(filename, "?", "？")
		filename = strings.ReplaceAll(filename, "*", "＊")
		filename = strings.TrimRight(filename, ". ")
	case "darwin":
		filename = strings.ReplaceAll(filename, ":", "：")
	}

	return strings.ReplaceAll(filename, "/", "／")
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

func autocropImages(images md.ImageList) {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", "Autocropping")
	bar.Set(pb.CleanOnFinish, true)
	bar.Start()
	bar.AddTotal(int64(len(images) - 1))
	defer bar.Finish()
	for i := 0; i < len(images); i++ {
		if cropped, err := autocrop(images[i].Image); err == nil {
			images[i].Image = cropped
		}
		bar.Add(-1)
	}
}

func autocrop(img image.Image) (image.Image, error) {
	configWhite := boundingbox.NewConfigInverse(color.Gray{200})
	configWhite.Parallel = boundingbox.ParallelAuto
	configBlack := boundingbox.NewConfig(color.Gray{50})
	configBlack.Parallel = boundingbox.ParallelAuto

	boxBlack := boundingbox.Find(img, configBlack)
	boxWhite := boundingbox.Find(img, configWhite)

	areaBlack := float64(boxBlack.Dx() * boxBlack.Dy())
	areaWhite := float64(boxWhite.Dx() * boxWhite.Dy())
	diff := ((math.Max(areaBlack, areaWhite) - math.Min(areaBlack, areaWhite)) / math.Min(areaBlack, areaWhite)) * 100
	if diff > 1 {
		if areaWhite > areaBlack {
			return cropImage(img, boxBlack)
		} else {
			return cropImage(img, boxWhite)
		}
	} else {
		return img, nil
	}
}

// cropImage takes an image and crops it to the specified rectangle.
func cropImage(img image.Image, crop image.Rectangle) (image.Image, error) {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	simg, ok := img.(subImager)
	if !ok {
		return nil, fmt.Errorf("image does not support cropping")
	}

	return simg.SubImage(crop), nil
}

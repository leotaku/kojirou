package download

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
	"go.uber.org/ratelimit"
	"golang.org/x/sync/errgroup"
)

const (
	maxChapterJobs = 8
	maxImageJobs   = 16
)

var (
	atHomeLimiter  ratelimit.Limiter = ratelimit.New(40, ratelimit.Per(time.Minute))
	httpClient     *http.Client
	mangadexClient *md.Client
)

func init() {
	retry := retryablehttp.NewClient()
	retry.Logger = nil
	httpClient = retry.StandardClient()
	mangadexClient = md.NewClient().WithHTTPClient(httpClient)
}

func MangadexSkeleton(mangaID string) (*md.Manga, error) {
	return mangadexClient.FetchManga(mangaID)

}

func MangadexChapters(mangaID string) (md.ChapterList, error) {
	return mangadexClient.FetchChapters(mangaID)
}

func MangadexCovers(manga *md.Manga, p formats.Progress) (md.ImageList, error) {
	covers, err := mangadexClient.FetchCovers(manga.Info.ID)
	if err != nil {
		return nil, err
	}

	pathQueue := make(chan md.Path, 100)
	imageQueue := make(chan md.Image, 100)
	go func() {
		for _, cover := range covers {
			if _, ok := manga.Volumes[cover.VolumeIdentifier]; ok {
				pathQueue <- cover
			}
		}
		close(pathQueue)
	}()

	eg := pathsToImages(pathQueue, imageQueue, context.TODO(), p)
	return collectImages(imageQueue, eg)
}

func MangadexPages(chapters md.ChapterList, p formats.Progress) (md.ImageList, error) {
	chapterQueue := make(chan md.Chapter, 10)
	pathQueue := make(chan md.Path, 100)
	pageQueue := make(chan md.Image, 100)

	ctx := context.TODO()
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(chaptersToPaths(chapterQueue, pathQueue, ctx, p).Wait)
	eg.Go(pathsToImages(pathQueue, pageQueue, ctx, p).Wait)

	go func() {
		for _, chapter := range chapters {
			chapterQueue <- chapter
		}
		close(chapterQueue)
	}()

	return collectImages(pageQueue, eg)
}

func collectImages(imageQueue <-chan md.Image, eg *errgroup.Group) (md.ImageList, error) {
	images := make(md.ImageList, 0)
	for image := range imageQueue {
		images = append(images, image)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	} else {
		return images, nil
	}
}

func chaptersToPaths(
	chapterQueue <-chan md.Chapter,
	pathQueue chan<- md.Path,
	ctx context.Context,
	progress formats.Progress,
) *errgroup.Group {
	return spinUp(ctx, maxChapterJobs, func() error {
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("cancelled")
			case chapter, running := <-chapterQueue:
				if !running {
					return nil
				}

				progress.Increase(1)
				atHomeLimiter.Take()
				paths, err := mangadexClient.FetchPaths(&chapter)
				if err != nil {
					return fmt.Errorf("chapter %v: paths: %w", chapter.Info.Identifier, err)
				}

				progress.Increase(len(paths) - 1)
				for _, path := range paths {
					pathQueue <- path
				}
			}
		}
	}, func() { close(pathQueue) })
}

func pathsToImages(
	pathQueue <-chan md.Path,
	imageQueue chan<- md.Image,
	ctx context.Context,
	progress formats.Progress,
) *errgroup.Group {
	return spinUp(ctx, maxImageJobs, func() error {
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("cancelled")
			case path, running := <-pathQueue:
				if !running {
					return nil
				}

				img, err := getImage(httpClient, path.URL)
				if err != nil {
					return fmt.Errorf("chapter %v: image %v: %w", path.ChapterIdentifier, path.ImageIdentifier, err)
				}

				progress.Add(1)
				imageQueue <- path.WithImage(img)
			}
		}
	}, func() { close(imageQueue) })
}

func getImage(client *http.Client, url string) (image.Image, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status: %v", resp.Status)
	}

	img, _, err := image.Decode(resp.Body)
	return img, err
}

func spinUp(ctx context.Context, concurrency int, f func() error, cleanup func()) *errgroup.Group {
	eg, _ := errgroup.WithContext(ctx)
	for i := 0; i < concurrency; i++ {
		eg.Go(f)
	}
	go func() {
		eg.Wait() //nolint:errcheck
		cleanup()
	}()

	return eg
}

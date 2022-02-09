package formats

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	md "github.com/leotaku/kojirou/mangadex"
	"go.uber.org/ratelimit"
	"golang.org/x/sync/errgroup"
)

const (
	maxChapterJobs = 8
	maxImageJobs   = 16
)

var (
	atHomeLimiter ratelimit.Limiter = ratelimit.New(40, ratelimit.Per(time.Minute))
)

type Reporter func(int)

type MangadexDownloader struct {
	Context  context.Context
	client   *md.Client
	http     *http.Client
	reporter Reporter
}

func NewMangadexDownloader(client *md.Client, http *http.Client, reporter Reporter) *MangadexDownloader {
	if reporter == nil {
		reporter = func(int) {}
	}

	return &MangadexDownloader{
		Context:  context.TODO(),
		client:   client,
		http:     http,
		reporter: reporter,
	}
}

func MangadexCovers(dl *MangadexDownloader, manga *md.Manga) (md.ImageList, error) {
	covers, err := dl.client.FetchCovers(manga.Info.ID)
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

	eg := dl.pathsToImages(pathQueue, imageQueue)
	return collectImages(imageQueue, eg)
}

func MangadexPages(dl *MangadexDownloader, chapters md.ChapterList) (md.ImageList, error) {
	chapterQueue := make(chan md.Chapter, 10)
	pathQueue := make(chan md.Path, 100)
	pageQueue := make(chan md.Image, 100)

	eg, _ := errgroup.WithContext(dl.Context)
	eg.Go(dl.chaptersToPaths(chapterQueue, pathQueue).Wait)
	eg.Go(dl.pathsToImages(pathQueue, pageQueue).Wait)

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

func (dl *MangadexDownloader) chaptersToPaths(
	chapterQueue <-chan md.Chapter,
	pathQueue chan<- md.Path,
) *errgroup.Group {
	return spinUp(dl.Context, maxChapterJobs, func() error {
		for {
			select {
			case <-dl.Context.Done():
				return fmt.Errorf("cancelled")
			case chapter, running := <-chapterQueue:
				if !running {
					return nil
				}

				dl.reporter(1)
				atHomeLimiter.Take()
				paths, err := dl.client.FetchPaths(&chapter)
				if err != nil {
					return fmt.Errorf("chapter %v: paths: %w", chapter.Info.Identifier, err)
				}

				dl.reporter(len(paths) - 1)
				for _, path := range paths {
					pathQueue <- path
				}
			}
		}
	}, func() { close(pathQueue) })
}

func (dl *MangadexDownloader) pathsToImages(
	pathQueue <-chan md.Path,
	imageQueue chan<- md.Image,
) *errgroup.Group {
	return spinUp(dl.Context, maxImageJobs, func() error {
		for {
			select {
			case <-dl.Context.Done():
				return fmt.Errorf("cancelled")
			case path, running := <-pathQueue:
				if !running {
					return nil
				}

				img, err := getImage(dl.http, path.URL)
				if err != nil {
					return fmt.Errorf("chapter %v: image %v: %w", path.ChapterIdentifier, path.ImageIdentifier, err)
				}

				dl.reporter(-1)
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

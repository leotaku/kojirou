package download

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/sync/errgroup"
)

type DataSaverPolicy int

const (
	DataSaverPolicyNo DataSaverPolicy = iota
	DataSaverPolicyPrefer
	DataSaverPolicyFallback
)

const (
	maxJobsChapter = 8
	maxJobsImage   = 16
)

var (
	httpClient     *http.Client
	mangadexClient *md.Client
)

func init() {
	retry := retryablehttp.NewClient()
	retry.Logger = nil
	retry.RetryWaitMin = time.Second * 5
	retry.Backoff = retryablehttp.LinearJitterBackoff
	retry.CheckRetry = bodyReadableErrorPolicy

	httpClient = retry.StandardClient()
	mangadexClient = md.NewClient().WithHTTPClient(httpClient)
}

func MangadexSkeleton(mangaID string) (*md.Manga, error) {
	return mangadexClient.FetchManga(context.TODO(), mangaID)
}

func MangadexChapters(mangaID string) (md.ChapterList, error) {
	return mangadexClient.FetchChapters(context.TODO(), mangaID)
}

func MangadexCovers(manga *md.Manga, saveRawArg bool, fillVolumeNumberArg int, p formats.Progress) (md.ImageList, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	covers, err := mangadexClient.FetchCovers(ctx, manga.Info.ID)
	if err != nil {
		return nil, err
	}

	coverPaths := make(chan md.Path)
	go func() {
		for _, path := range covers {
			if _, ok := manga.Volumes[path.VolumeIdentifier]; ok {
				coverPaths <- path
				p.Increase(1)
			}
		}
		close(coverPaths)
	}()

	coverImages, eg := pathsToImages(coverPaths, ctx, cancel, DataSaverPolicyNo, saveRawArg, fillVolumeNumberArg, manga.Info.Title)

	results := make(md.ImageList, len(covers))
	for coverImage := range coverImages {
		p.Add(1)
		results = append(results, coverImage)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	} else {
		return results, nil
	}
}

func MangadexPages(
	chapterList md.ChapterList,
	policy DataSaverPolicy,
	saveRawArg bool,
	fillVolumeNumberArg int,
	mangaTitle string,
	p formats.Progress,
) (md.ImageList, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	chapters := make(chan md.Chapter)
	go func() {
		for _, chapter := range chapterList {
			chapters <- chapter
			p.Increase(1)
		}
		close(chapters)
	}()

	paths, childEg := chaptersToPaths(chapters, ctx, cancel, p)
	eg.Go(childEg.Wait)

	images, childEg := pathsToImages(paths, ctx, cancel, policy, saveRawArg, fillVolumeNumberArg, mangaTitle)
	eg.Go(childEg.Wait)

	results := make(md.ImageList, 0)
	for image := range images {
		p.Add(1)
		results = append(results, image)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	} else {
		return results, nil
	}
}

func chaptersToPaths(
	chapters <-chan md.Chapter,
	ctx context.Context,
	cancel context.CancelFunc,
	p formats.Progress,
) (<-chan md.Path, *errgroup.Group) {
	ch := make(chan md.Path)
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(maxJobsChapter + 1)

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("canceled")
			case chapter, ok := <-chapters:
				if !ok {
					return nil
				}
				eg.Go(func() error {
					paths, err := mangadexClient.FetchPaths(ctx, &chapter)
					if err != nil {
						defer cancel()
						return fmt.Errorf("chapter %v: paths: %w", chapter.Info.Identifier, err)
					} else {
						p.Add(1)
						for _, path := range paths {
							select {
							case <-ctx.Done():
								return fmt.Errorf("canceled")
							case ch <- path:
								p.Increase(1)
							}
						}
						return nil
					}
				})
			}
		}
	})

	go func() {
		eg.Wait() //nolint:errcheck
		close(ch)
	}()

	return ch, eg
}

func pathsToImages(
	paths <-chan md.Path,
	ctx context.Context,
	cancel context.CancelFunc,
	policy DataSaverPolicy,
	saveRawArg bool,
	fillVolumeNumberArg int,
	mangaTitle string,
) (<-chan md.Image, *errgroup.Group) {
	ch := make(chan md.Image)
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(maxJobsImage + 1)

	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("canceled")
			case path, ok := <-paths:
				if !ok {
					return nil
				}
				eg.Go(func() error {
					img, err := getImageWithPolicy(httpClient, ctx, path, policy, saveRawArg, fillVolumeNumberArg, mangaTitle)
					if err != nil {
						defer cancel()
						return fmt.Errorf("chapter %v: image %v: %w", path.ChapterIdentifier, path.ImageIdentifier, err)
					}

					select {
					case <-ctx.Done():
						return fmt.Errorf("canceled")
					case ch <- path.WithImage(img):
						return nil
					}
				})
			}
		}
	})

	go func() {
		eg.Wait() //nolint:errcheck
		close(ch)
	}()

	return ch, eg
}

func getImageWithPolicy(
	client *http.Client,
	ctx context.Context,
	path md.Path,
	policy DataSaverPolicy,
	saveRawArg bool,
	fillVolumeNumberArg int,
	mangaTitle string,
) (image.Image, error) {
	resp := new(http.Response)
	err := error(nil)

	switch policy {
	case DataSaverPolicyNo, DataSaverPolicyFallback:
		resp, err = getResp(client, ctx, path.DataURL)
	case DataSaverPolicyPrefer:
		resp, err = getResp(client, ctx, path.DataSaverURL)
	}

	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	img, _, err := image.Decode(resp.Body)
	defer resp.Body.Close()

	if err != nil && policy == DataSaverPolicyFallback {
		return getImageWithPolicy(client, ctx, path, DataSaverPolicyPrefer, saveRawArg, fillVolumeNumberArg, mangaTitle)
	} else if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if saveRawArg {
		// Save the image to a temporary directory
		tempDir := filepath.Join("raw_images", mangaTitle, "Volume "+path.VolumeIdentifier.StringFilled(fillVolumeNumberArg, 0, false))
		// if chapter id & img id are both 0, it's a cover image
		if !(path.ChapterIdentifier.String() == "0" && path.ImageIdentifier == 0) {
			tempDir = filepath.Join(tempDir, fmt.Sprintf("Chapter %s", path.ChapterIdentifier))
		}
		if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
		tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%d.jpg", path.ImageIdentifier))
		if path.ChapterIdentifier.String() == "0" && path.ImageIdentifier == 0 {
			tempFilePath = filepath.Join(tempDir, "cover.jpg")
		}
		tempFile, err := os.Create(tempFilePath)
		if err != nil {
			return nil, fmt.Errorf("create temp file: %w", err)
		}
		defer tempFile.Close()

		if err := jpeg.Encode(tempFile, img, nil); err != nil {
			return nil, fmt.Errorf("save temp image: %w", err)
		}
	}

	return img, nil
}

func getResp(client *http.Client, ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status: %v", resp.Status)
	}

	return resp, nil
}

func bodyReadableErrorPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if retry, err := retryablehttp.DefaultRetryPolicy(ctx, resp, err); retry || err != nil {
		return retry, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(buf)

	if err != nil {
		return true, nil
	} else {
		return false, nil
	}
}

package download

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/leotaku/kojirou/cmd/formats"
	md "github.com/leotaku/kojirou/mangadex"
	"golang.org/x/sync/errgroup"
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

func MangadexCovers(manga *md.Manga, p formats.Progress) (md.ImageList, error) {
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

	coverImages, eg := pathsToImages(coverPaths, ctx, cancel, DataSaverPolicyNo)

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

func MangadexPages(chapterList md.ChapterList, policy DataSaverPolicy, p formats.Progress) (md.ImageList, error) {
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

	images, childEg := pathsToImages(paths, ctx, cancel, policy)
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
					img, err := getImageWithPolicy(httpClient, ctx, path, policy)
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

func getImageWithPolicy(client *http.Client, ctx context.Context, path md.Path, policy DataSaverPolicy) (image.Image, error) {
	resp := new(http.Response)
	err := error(nil)

	switch policy {
	case DataSaverPolicyNo, DataSaverPolicyFallback:
		resp, err = getResp(httpClient, ctx, path.DataURL)
	case DataSaverPolicyPrefer:
		resp, err = getResp(httpClient, ctx, path.DataSaverURL)
	}

	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	img, _, err := image.Decode(resp.Body)
	defer resp.Body.Close()

	if err != nil && policy == DataSaverPolicyFallback {
		return getImageWithPolicy(client, ctx, path, DataSaverPolicyPrefer)
	} else if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	} else {
		return img, nil
	}
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

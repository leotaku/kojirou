package util

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/leotaku/kojirou/mangadex"
)

const (
	maxChapterJobs = 8
	maxImageJobs   = 16
)

var (
	Client     *mangadex.Client
	httpClient *http.Client
)

type pathOrErr struct {
	page mangadex.PathItem
	err  error
}

type imageOrErr struct {
	page mangadex.ImageItem
	err  error
}

func FetchPages(cs mangadex.ChapterList, pb *Bar) (mangadex.ImageList, error) {
	pb.AddTotal(int64(len(cs)))

	// Fetch chapters in parallel
	wip := make(chan pathOrErr, 200)
	go runLinks(cs, wip, pb)

	// Fetch images in parallel
	images, err := fetchImages(wip, pb)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func FetchCovers(ps mangadex.PathList, pb *Bar) (mangadex.ImageList, error) {
	pb.AddTotal(int64(len(ps)))
	in := make(chan pathOrErr)
	go func() {
		for _, path := range ps {
			in <- pathOrErr{page: path}
		}
		close(in)
	}()

	return fetchImages(in, pb)
}

func fetchImages(in <-chan pathOrErr, pb *Bar) (mangadex.ImageList, error) {
	result := make(mangadex.ImageList, 0)
	for it := range runImages(in, pb) {
		if it.err != nil {
			return nil, it.err
		}
		result = append(result, it.page)
	}
	return result, nil
}

func runLinks(chaps []mangadex.ChapterInfo, out chan pathOrErr, pb *Bar) {
	in := make(chan mangadex.ChapterInfo)
	wg := new(sync.WaitGroup)

	for i := 0; i < maxChapterJobs; i++ {
		wg.Add(1)
		go func() {
			for it := range in {
				paths, err := Client.FetchPaths(&it)
				pb.Increment()
				pb.AddTotal(int64(len(paths)))
				for _, path := range paths {
					out <- pathOrErr{page: path}
				}

				if err != nil {
					out <- pathOrErr{err: err}
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	go func() {
		for _, chap := range chaps {
			in <- chap
		}
		close(in)
	}()
}

func runImages(in <-chan pathOrErr, pb *Bar) chan imageOrErr {
	out := make(chan imageOrErr, 100)
	wg := new(sync.WaitGroup)

	for i := 0; i < maxImageJobs; i++ {
		wg.Add(1)
		go func() {
			for it := range in {
				if it.err != nil {
					out <- imageOrErr{err: it.err}
					return
				}

				img, err := fetchImage(it.page.URL)
				pb.Increment()

				if err != nil {
					out <- imageOrErr{err: err}
				} else {
					out <- imageOrErr{page: it.page.WithImage(img)}
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func fetchImage(url string) (image.Image, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status: %v", resp.StatusCode)
	} else if img, _, err := image.Decode(resp.Body); err != nil {
		return nil, fmt.Errorf("decode: %v", err)
	} else {
		return img, nil
	}
}

func init() {
	retry := retryablehttp.NewClient()
	retry.Logger = nil
	httpClient = retry.StandardClient()
	Client = mangadex.NewClient().WithHTTPClient(*httpClient)
}

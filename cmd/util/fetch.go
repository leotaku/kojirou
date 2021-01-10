package util

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/leotaku/manki/mangadex"
)

const limitArg = 8

var (
	Client     *mangadex.Client
	httpClient *http.Client
)

func init() {
	retry := retryablehttp.NewClient()
	retry.Logger = nil
	httpClient = retry.StandardClient()
	Client = mangadex.NewClient().WithHTTPClient(*httpClient)
}

type unitem struct {
	page mangadex.PathItem
	err  error
}

type item struct {
	page mangadex.ImageItem
	err  error
}

func FetchChapters(cs mangadex.ChapterList, pb *Bar) (mangadex.ImageList, error) {
	pb.AddTotal(int64(len(cs)))

	// Fetch chapters in parallel
	wip := make(chan unitem, 200)
	go runChapters(cs, wip, pb)

	// Fetch images in parallel
	images, err := fetchImages(wip, pb)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func FetchCovers(cs mangadex.PathList, pb *Bar) (mangadex.ImageList, error) {
	pb.AddTotal(int64(len(cs)))
	in := make(chan unitem)
	go func() {
		for _, path := range cs {
			in <- unitem{page: path}
		}
		close(in)
	}()

	return fetchImages(in, pb)
}

func fetchImages(in <-chan unitem, pb *Bar) (mangadex.ImageList, error) {
	result := make(mangadex.ImageList, 0)
	for it := range runImages(in, pb) {
		if it.err != nil {
			return nil, it.err
		}
		result = append(result, it.page)
	}
	return result, nil
}

func runChapters(chaps []mangadex.ChapterInfo, out chan unitem, pb *Bar) {
	in := make(chan mangadex.ChapterInfo)
	wg := new(sync.WaitGroup)

	for i := 0; i < limitArg; i++ {
		wg.Add(1)
		go func() {
			for it := range in {
				paths, err := Client.FetchChapter(it)
				pb.Increment()
				pb.AddTotal(int64(len(paths)))
				for _, path := range paths {
					out <- unitem{page: path}
				}

				if err != nil {
					out <- unitem{err: err}
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

func runImages(in <-chan unitem, pb *Bar) chan item {
	out := make(chan item, 100)
	wg := new(sync.WaitGroup)

	for i := 0; i < limitArg; i++ {
		wg.Add(1)
		go func() {
			for it := range in {
				if it.err != nil {
					out <- item{err: it.err}
					return
				}

				img, err := fetchImage(it.page.URL)
				pb.Increment()

				if err != nil {
					out <- item{err: err}
				} else {
					out <- item{page: it.page.WithImage(img)}
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

	img, _, err := image.Decode(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return img, err
}

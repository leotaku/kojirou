package util

import (
	"fmt"
	"sync"

	"github.com/leotaku/manki/mangadex"
)

const limitArg = 8

type item struct {
	pages []mangadex.PathInfo
	err   error
}

type page struct {
	page mangadex.ImageInfo
	err  error
}

func FetchChapters(cs []mangadex.ChapterInfo, pb *Bar) ([]mangadex.ImageInfo, error) {
	pb.AddTotal(int64(len(cs)))

	// Fetch chapters in parallel
	wip := runChapters(cs)
	pres := make([]mangadex.PathInfo, 0)
	for it := range wip {
		if it.err == nil {
			pres = append(pres, it.pages...)
			pb.AddTotal(int64(len(it.pages)))
			pb.Increment()
		} else {
			pb.Fail("Failed fetching chapters")
			return nil, fmt.Errorf("Chapters: %w", it.err)
		}
	}

	// Fetch images in parallel
	pages, err := fetchImages(pres, pb)
	if err != nil {
		return nil, err
	}

	return pages, nil
}

func FetchCovers(cs []mangadex.PathInfo, pb *Bar) ([]mangadex.ImageInfo, error) {
	pb.AddTotal(int64(len(cs)))
	return fetchImages(cs, pb)
}

func fetchImages(cs []mangadex.PathInfo, pb *Bar) ([]mangadex.ImageInfo, error) {
	out := runImages(cs)
	covers := make([]mangadex.ImageInfo, 0)
	for it := range out {
		if it.err == nil {
			covers = append(covers, it.page)
			pb.Increment()
		} else {
			pb.Fail("Failed fetching images")
			return nil, fmt.Errorf("Images: %w", it.err)
		}
	}

	return covers, nil
}

func runImages(pages []mangadex.PathInfo) chan page {
	in := make(chan mangadex.PathInfo)
	out := make(chan page)
	wg := new(sync.WaitGroup)

	for i := 0; i < limitArg; i++ {
		wg.Add(1)
		go func() {
			for it := range in {
				err := Retry(func() error {
					paths, err := it.GetImage()
					if err == nil {
						out <- page{page: *paths}
					}
					return err
				})

				if err != nil {
					out <- page{err: err}
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
		for _, it := range pages {
			in <- it
		}
		close(in)
	}()

	return out
}

func runChapters(chaps []mangadex.ChapterInfo) chan item {
	in := make(chan mangadex.ChapterInfo)
	out := make(chan item)
	wg := new(sync.WaitGroup)

	for i := 0; i < limitArg; i++ {
		wg.Add(1)
		go func() {
			for it := range in {
				err := Retry(func() error {
					paths, err := it.GetImagePaths()
					if err == nil {
						out <- item{pages: paths}
					}
					return err
				})

				if err != nil {
					out <- item{err: err}
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

	return out
}

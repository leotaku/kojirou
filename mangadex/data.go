package mangadex

import (
	"fmt"
	"image"
	"sync"

	"github.com/leotaku/manki/mangadex/api"
	"golang.org/x/text/language"
)

type Manga struct {
	Title       string
	Authors     []string
	Artists     []string
	Description string
	IsHentai    bool
	Id          int
	Versions    []Version
}

type Version struct {
	GroupNames []string
	Volumes    []Volume
	Missing    []Identifier
	Region     language.Region
}

type Volume struct {
	Number     Identifier
	CoverImage Image
	Chapters   []Chapter
}

type Chapter struct {
	Title  string
	Number Identifier
	Views  int
	Hash   string
	Id     int
	Images []Image `json:",omitempty"`
}

type Image struct {
	Url   string
	Image image.Image
}

func Fetch(mangaID int) (*Manga, error) {
	b, err := api.FetchBase(mangaID)
	if err != nil {
		return nil, fmt.Errorf("Fetch base: %w", err)
	}
	co, err := api.FetchCovers(mangaID)
	if err != nil {
		return nil, fmt.Errorf("Fetch covers: %w", err)
	}
	ca, err := api.FetchChapters(mangaID)
	if err != nil {
		return nil, fmt.Errorf("Fetch chapters: %w", err)
	}

	manga := convert(b.Data, ca.Data, co.Data)
	return &manga, nil
}

func FetchAsync(mangaID int) (*Manga, error) {
	wg := new(sync.WaitGroup)
	wg.Add(3)

	// Fetch base data
	var b *api.Base
	var be error
	go func() {
		b, be = api.FetchBase(mangaID)
		wg.Done()
	}()

	// Fetch chapter data
	var ca *api.Chapters
	var cae error
	go func() {
		ca, cae = api.FetchChapters(mangaID)
		wg.Done()
	}()

	// Fetch cover data
	var co *api.Covers
	var coe error
	go func() {
		co, coe = api.FetchCovers(mangaID)
		wg.Done()
	}()

	wg.Wait()
	if be != nil || cae != nil || coe != nil {
		return nil, fmt.Errorf("%v, %v, %v", be, cae, coe)
	}

	manga := convert(b.Data, ca.Data, co.Data)
	return &manga, nil
}

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const ApiBaseURL = `https://mangadex.org/api/v2/`

type Base struct {
	Code   int
	Status string
	Data   BaseData
}

type BaseData struct {
	Id           int
	Title        string
	AltTitles    []string
	Description  string
	Artist       []string
	Author       []string
	Tags         []int
	LastChapter  string
	LastVolume   string
	IsHentai     bool
	Links        map[string]string
	Relations    []Relation
	Views        int
	Follows      int
	Comments     int
	LastUploaded int
	MainCover    string
	Publication  struct {
		Language    string
		Status      int
		Demographic int
	}
	Rating struct {
		Bayesian float64
		Mean     float64
		Users    int
	}
}

type Relation struct {
	Id       int
	Title    string
	Type     int
	IsHentai bool
}

type Chapters struct {
	Code   int
	Status string
	Data   ChaptersData
}

type ChaptersData struct {
	Chapters []ChapterInfo
	Groups   []GroupMapping
}

type ChapterInfo struct {
	Id         int
	Hash       string
	MangaId    int
	MangaTitle string
	Volume     string
	Chapter    string
	Title      string
	Language   string
	Groups     []int
	Uploader   int
	Timestamp  int
	Comments   int
	Views      int
}

type GroupMapping struct {
	Id   int
	Name string
}

type Chapter struct {
	Code   int
	Status string
	Data   ChapterData
}

type ChapterData struct {
	Id         int
	Hash       string
	MangaId    int
	MangaTitle string
	Volume     string
	Chapter    string
	Title      string
	Language   string
	Groups     []GroupMapping
	Uploader   int
	Timestamp  int
	Comments   int
	Views      int
	Status     string
	Pages      []string
	Server     string
}

type Covers struct {
	Code   int
	Status string
	Data   CoversData
}

type CoversData []CoversMapping

type CoversMapping struct {
	Volume string
	Url    string
}

func FetchBase(mangaID int) (*Base, error) {
	v := new(Base)
	err := fetchJSON(v, "%v/manga/%v", ApiBaseURL, mangaID)
	return v, err
}

func FetchChapters(mangaID int) (*Chapters, error) {
	v := new(Chapters)
	err := fetchJSON(v, "%v/manga/%v/chapters", ApiBaseURL, mangaID)
	return v, err
}

func FetchCovers(mangaID int) (*Covers, error) {
	v := new(Covers)
	err := fetchJSON(v, "%v/manga/%v/covers", ApiBaseURL, mangaID)
	return v, err
}

func FetchChapter(chapterID int) (*Chapter, error) {
	v := new(Chapter)
	err := fetchJSON(v, "%v/chapter/%v", ApiBaseURL, chapterID)
	return v, err
}

type HttpStatusError struct {
	status int
}

func (e HttpStatusError) Status() int {
	return e.status
}

func (e HttpStatusError) Error() string {
	return fmt.Sprintf("Http status code: %v", e.Status())
}

func fetchJSON(v interface{}, url string, a ...interface{}) error {
	resp, err := http.Get(fmt.Sprintf(url, a...))
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return HttpStatusError{status: resp.StatusCode}
	}

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	dec.DisallowUnknownFields()

	err = dec.Decode(v)
	if err != nil {
		return err
	}

	return nil
}

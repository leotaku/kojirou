package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const ApiBaseURL = `https://mangadex.org/api/v2/`

type Client struct {
	Inner   http.Client
	BaseUrl string
}

func NewClient() *Client {
	return &Client{
		Inner:   *http.DefaultClient,
		BaseUrl: ApiBaseURL,
	}
}

func (c *Client) WithBaseURL(url string) *Client {
	c.BaseUrl = url
	return c
}

func (c *Client) WithClient(http http.Client) *Client {
	c.Inner = http
	return c
}

func (c *Client) FetchBase(mangaID int) (*Base, error) {
	v := new(Base)
	err := c.fetchJSON(v, "%v/manga/%v", ApiBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchChapters(mangaID int) (*Chapters, error) {
	v := new(Chapters)
	err := c.fetchJSON(v, "%v/manga/%v/chapters", ApiBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchCovers(mangaID int) (*Covers, error) {
	v := new(Covers)
	err := c.fetchJSON(v, "%v/manga/%v/covers", ApiBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchChapter(chapterID int) (*Chapter, error) {
	v := new(Chapter)
	err := c.fetchJSON(v, "%v/chapter/%v", ApiBaseURL, chapterID)
	return v, err
}

func (c *Client) fetchJSON(v interface{}, url string, a ...interface{}) error {
	resp, err := c.Inner.Get(fmt.Sprintf(url, a...))
	if err != nil {
		return err
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

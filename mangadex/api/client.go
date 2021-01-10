package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const APIBaseURL = `https://mangadex.org/api/v2/`

type Client struct {
	Inner   http.Client
	BaseURL string
}

func NewClient() *Client {
	return &Client{
		Inner:   *http.DefaultClient,
		BaseURL: APIBaseURL,
	}
}

func (c *Client) WithBaseURL(url string) *Client {
	c.BaseURL = url
	return c
}

func (c *Client) WithClient(http http.Client) *Client {
	c.Inner = http
	return c
}

func (c *Client) FetchBase(mangaID int) (*Base, error) {
	v := new(Base)
	err := c.fetchJSON(v, "%v/manga/%v", APIBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchChapters(mangaID int) (*Chapters, error) {
	v := new(Chapters)
	err := c.fetchJSON(v, "%v/manga/%v/chapters", APIBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchCovers(mangaID int) (*Covers, error) {
	v := new(Covers)
	err := c.fetchJSON(v, "%v/manga/%v/covers", APIBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchChapter(chapterID int) (*Chapter, error) {
	v := new(Chapter)
	err := c.fetchJSON(v, "%v/chapter/%v", APIBaseURL, chapterID)
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

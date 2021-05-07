package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const APIBaseURL = `https://api.mangadex.org`

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

func (c *Client) FetchBase(mangaID string) (*Base, error) {
	v := new(Base)
	err := c.fetchJSON(v, "%v/manga/%v", APIBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchFeed(mangaID string) (*Feed, error) {
	v := new(Feed)
	err := c.fetchJSON(v, "%v/manga/%v/feed?limit=500", APIBaseURL, mangaID)
	return v, err
}

func (c *Client) FetchChapter(chapterID string) (*Chapter, error) {
	v := new(Chapter)
	err := c.fetchJSON(v, "%v/chapter/%v", APIBaseURL, chapterID)
	return v, err
}

func (c *Client) FetchCreator(creatorID string) (*Creator, error) {
	v := new(Creator)
	err := c.fetchJSON(v, "%v/author/%v", APIBaseURL, creatorID)
	return v, err
}

func (c *Client) FetchGroup(groupID string) (*Group, error) {
	v := new(Group)
	err := c.fetchJSON(v, "%v/group/%v", APIBaseURL, groupID)
	return v, err
}

func (c *Client) FetchAtHome(chapterID string) (*AtHome, error) {
	v := new(AtHome)
	err := c.fetchJSON(v, "%v/at-home/server/%v", APIBaseURL, chapterID)
	return v, err
}


func (c *Client) fetchJSON(v interface{}, url string, a ...interface{}) error {
	resp, err := c.Inner.Get(fmt.Sprintf(url, a...))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	if 	err := dec.Decode(v); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}

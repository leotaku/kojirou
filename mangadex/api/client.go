package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var APIBaseURL, _ = url.Parse(`https://api.mangadex.org/`)

type Client struct {
	Inner   http.Client
	BaseURL url.URL
}

func NewClient() *Client {
	return &Client{
		Inner:   *http.DefaultClient,
		BaseURL: *APIBaseURL,
	}
}

func (c *Client) WithBaseURL(url url.URL) *Client {
	c.BaseURL = url
	return c
}

func (c *Client) WithClient(http http.Client) *Client {
	c.Inner = http
	return c
}

func (c *Client) GetManga(mangaID string) (*Manga, error) {
	v := new(Manga)
	err := c.doJSON("GET", "/manga/"+mangaID, v, nil)
	return v, err
}

func (c *Client) GetFeed(mangaID string, args QueryArgs) (*ChapterList, error) {
	v := new(ChapterList)
	url := fmt.Sprintf("/manga/%v/feed?%v", mangaID, args.Values().Encode())
	err := c.doJSON("GET", url, v, nil)
	return v, err
}

func (c *Client) GetCovers(args QueryArgs) (*CoverList, error) {
	v := new(CoverList)
	err := c.doJSON("GET", "/cover?"+args.Values().Encode(), v, nil)
	return v, err
}

func (c *Client) GetAuthors(args QueryArgs) (*AuthorList, error) {
	v := new(AuthorList)
	err := c.doJSON("GET", "/author?"+args.Values().Encode(), v, nil)
	return v, err
}

func (c *Client) GetGroups(args QueryArgs) (*GroupList, error) {
	v := new(GroupList)
	err := c.doJSON("GET", "/group?"+args.Values().Encode(), v, nil)
	return v, err
}

func (c *Client) GetAtHome(chapterID string) (*AtHome, error) {
	v := new(AtHome)
	err := c.doJSON("GET", "/at-home/server/"+chapterID, v, nil)
	return v, err
}

func (c *Client) PostIDMapping(tp string, legacyIDs ...int) ([]IDMapping, error) {
	v := make([]IDMapping, 0)
	err := c.doJSON("POST", "/legacy/mapping", &v, map[string]interface{}{
		"ids":  legacyIDs,
		"type": tp,
	})

	return v, err
}

func (c *Client) doJSON(method, ref string, result, body interface{}) error {
	url, err := c.BaseURL.Parse(ref)
	if err != nil {
		return fmt.Errorf("url: %w", err)
	}

	rw := io.ReadWriter(nil)
	if body != nil {
		rw = bytes.NewBuffer(nil)
		if err := json.NewEncoder(rw).Encode(body); err != nil {
			return fmt.Errorf("encode: %w", err)
		}
	}

	req, err := http.NewRequest(method, url.String(), rw)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Inner.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errs := new(Errors)
		if err := dec.Decode(err); len(errs.Errors) != 0 {
			return fmt.Errorf(errs.Errors[0].Detail)
		} else {
			return fmt.Errorf("decode error: %w", err)
		}
	} else if err := dec.Decode(result); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/ratelimit"
)

var (
	limitGlobal = ratelimit.New(5, ratelimit.Per(time.Second))
	limitAtHome = ratelimit.New(40, ratelimit.Per(time.Minute))
)

var APIBaseURL, _ = url.Parse(`https://api.mangadex.org/`)

type Client struct {
	http    *http.Client
	baseURL url.URL
}

func NewClient() *Client {
	return &Client{
		http:    http.DefaultClient,
		baseURL: *APIBaseURL,
	}
}

func (c *Client) WithBaseURL(url url.URL) *Client {
	c.baseURL = url
	return c
}

func (c *Client) WithHTTPClient(http *http.Client) *Client {
	c.http = http
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
	limitAtHome.Take()
	err := c.doJSON("GET", "/at-home/server/"+chapterID, v, nil)
	return v, err
}

func (c *Client) PostIDMapping(tp string, legacyIDs ...int) (*IDMappingList, error) {
	v := new(IDMappingList)
	err := c.doJSON("POST", "/legacy/mapping", &v, map[string]interface{}{
		"ids":  legacyIDs,
		"type": tp,
	})

	return v, err
}

func (c *Client) doJSON(method, ref string, result, body interface{}) error {
	url, err := c.baseURL.Parse(ref)
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

	limitGlobal.Take()
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errs := new(Errors)
		if err := dec.Decode(errs); err != nil {
			return fmt.Errorf("error decode: %w", err)
		} else if len(errs.Errors) != 0 {
			return fmt.Errorf("detail: %s", errs.Errors[0].Detail)
		} else {
			return fmt.Errorf("status: %v", resp.Status)
		}
	} else if err := dec.Decode(result); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}

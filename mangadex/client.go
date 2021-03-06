package mangadex

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/leotaku/kojirou/mangadex/api"
)

var CoverBaseURL, _ = url.Parse("https://uploads.mangadex.org/covers/")

type Client struct {
	base         api.Client
	coverBaseURL url.URL
}

func NewClient() *Client {
	return &Client{
		base:         *api.NewClient(),
		coverBaseURL: *CoverBaseURL,
	}
}

func (c *Client) WithHTTPClient(http http.Client) *Client {
	c.base.WithClient(http)
	return c
}

func (c *Client) FetchLegacy(tp string, legacyID int) (api.StringID, error) {
	ids, err := c.base.PostIDMapping(tp, legacyID)
	if err != nil {
		return "", fmt.Errorf("post mapping: %w", err)
	}

	if len(ids) != 1 {
		return "", fmt.Errorf("%v not found: %v", tp, legacyID)
	}

	return ids[0].Data.Attributes.NewID, nil
}

func (c *Client) FetchManga(mangaID string) (*Manga, error) {
	b, err := c.base.GetManga(mangaID)
	if err != nil {
		return nil, fmt.Errorf("get manga: %w", err)
	}

	// Only retrieves at most 100 authors
	authors, err := c.base.GetAuthors(api.QueryArgs{
		IDs:   b.Relationships.Author,
		Limit: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("get authors: %w", err)
	}

	// Only retrieves at most 100 artists
	artists, err := c.base.GetAuthors(api.QueryArgs{
		IDs:   b.Relationships.Artist,
		Limit: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("get artists: %w", err)
	}

	return &Manga{
		Info:    convertManga(b, authors, artists),
		Volumes: make(map[Identifier]Volume),
	}, nil
}

func (c *Client) FetchChapters(mangaID string) (ChapterList, error) {
	chapters := make([]api.Chapter, 0)

	limit := 500
	for offset := 0; ; offset += limit {
		feed, err := c.base.GetFeed(mangaID, api.QueryArgs{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("get chapters: %w", err)
		} else {
			chapters = append(chapters, feed.Results...)
		}

		if offset+limit >= feed.Total {
			break
		}
	}

	groupMap, err := c.fetchGroupMap(chapters)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}

	return convertChapters(chapters, groupMap), nil
}

func (c *Client) FetchCovers(mangaID string) (PathList, error) {
	covers := make([]api.Cover, 0)
	limit := 100
	for offset := 0; ; offset += limit {
		feed, err := c.base.GetCovers(api.QueryArgs{
			Mangas: []string{mangaID},
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("get covers: %w", err)
		} else {
			covers = append(covers, feed.Results...)
		}

		if offset+limit >= feed.Total {
			break
		}
	}

	return convertCovers(c.coverBaseURL.String(), mangaID, covers), nil
}

func (c *Client) FetchPaths(chapter *ChapterInfo) (PathList, error) {
	ah, err := c.base.GetAtHome(chapter.ID)
	if err != nil {
		return nil, fmt.Errorf("get at home: %w", err)
	}

	return convertChapter(ah.BaseURL, chapter), nil
}

func (c *Client) fetchGroupMap(chapters []api.Chapter) (map[string]api.Group, error) {
	dedup := make(map[string]struct{})
	groupIDs := make([]string, 0)
	for _, chap := range chapters {
		for _, id := range chap.Relationships.Group {
			if _, ok := dedup[id]; !ok {
				groupIDs = append(groupIDs, id)
				dedup[id] = struct{}{}
			}
		}
	}

	result := make(map[string]api.Group)
	limit := 100
	for offset := 0; ; offset += limit {
		gs, err := c.base.GetGroups(api.QueryArgs{
			IDs:    groupIDs,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("fetch chapters: %w", err)
		} else {
			for _, group := range gs.Results {
				result[group.Data.ID] = group
			}
		}

		if offset+limit >= gs.Total {
			break
		}
	}

	return result, nil
}

package mangadex

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/leotaku/kojirou/mangadex/api"
)

var CoverBaseURL, _ = url.Parse("https://uploads.mangadex.org/covers/")

type Client struct {
	base         *api.Client
	coverBaseURL url.URL
}

func NewClient() *Client {
	return &Client{
		base:         api.NewClient(),
		coverBaseURL: *CoverBaseURL,
	}
}

func (c *Client) WithHTTPClient(http *http.Client) *Client {
	c.base.WithHTTPClient(http)
	return c
}

func (c *Client) FetchLegacy(ctx context.Context, tp string, legacyID int) (string, error) {
	mapping, err := c.base.PostIDMapping(ctx, tp, legacyID)
	if err != nil {
		return "", fmt.Errorf("post mapping: %w", err)
	}

	if len(mapping.Data) != 1 {
		return "", fmt.Errorf("%v not found: %v", tp, legacyID)
	}

	return mapping.Data[0].Attributes.NewID, nil
}

func (c *Client) FetchManga(ctx context.Context, mangaID string) (*Manga, error) {
	base, err := c.base.GetManga(ctx, mangaID)
	if err != nil {
		return nil, fmt.Errorf("get manga: %w", err)
	}

	// Only retrieves at most 100 authors
	authors, err := c.base.GetAuthors(ctx, api.QueryArgs{
		IDs:   base.Data.Relationships.Author,
		Limit: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("get authors: %w", err)
	}

	// Only retrieves at most 100 artists
	artists, err := c.base.GetAuthors(ctx, api.QueryArgs{
		IDs:   base.Data.Relationships.Artist,
		Limit: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("get artists: %w", err)
	}

	return &Manga{
		Info:    convertManga(base, authors, artists),
		Volumes: make(map[Identifier]Volume),
	}, nil
}

func (c *Client) FetchChapters(ctx context.Context, mangaID string) (ChapterList, error) {
	chapters := make([]api.ChapterData, 0)

	limit := 500
	for offset := 0; ; offset += limit {
		feed, err := c.base.GetFeed(ctx, mangaID, api.QueryArgs{
			Limit:         limit,
			Offset:        offset,
			Order:         map[string]string{"updatedAt": "asc"},
			EmptyPages:    "0",
			FuturePublish: "0",
			ExternalURL:   "0",
		})
		if err != nil {
			return nil, fmt.Errorf("get chapters: %w", err)
		} else {
			chapters = append(chapters, feed.Data...)
		}

		if offset+limit >= feed.Total {
			break
		}
	}

	groupMap, err := c.fetchGroupMap(ctx, chapters)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}

	return convertChapters(chapters, groupMap), nil
}

func (c *Client) FetchCovers(ctx context.Context, mangaID string) (PathList, error) {
	covers := make([]api.CoverData, 0)
	limit := 100
	for offset := 0; ; offset += limit {
		feed, err := c.base.GetCovers(ctx, api.QueryArgs{
			Mangas: []string{mangaID},
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("get covers: %w", err)
		} else {
			covers = append(covers, feed.Data...)
		}

		if offset+limit >= feed.Total {
			break
		}
	}

	return convertCovers(c.coverBaseURL.String(), mangaID, covers), nil
}

func (c *Client) FetchPaths(ctx context.Context, chapter *Chapter) (PathList, error) {
	ah, err := c.base.GetAtHome(ctx, chapter.Info.ID)
	if err != nil {
		return nil, fmt.Errorf("get at home: %w", err)
	}

	return convertChapter(chapter, ah), nil
}

func (c *Client) fetchGroupMap(ctx context.Context, chapters []api.ChapterData) (map[string]api.GroupData, error) {
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

	result := make(map[string]api.GroupData)
	limit := 100
	for offset := 0; offset < len(groupIDs); offset += limit {
		// Always send at most `limit` IDs
		end := len(groupIDs)
		if end > offset+limit {
			end = offset + limit
		}

		gs, err := c.base.GetGroups(ctx, api.QueryArgs{
			IDs:   groupIDs[offset:end],
			Limit: limit,
		})
		if err != nil {
			return nil, err
		} else {
			for _, group := range gs.Data {
				result[group.ID] = group
			}
		}
	}

	return result, nil
}

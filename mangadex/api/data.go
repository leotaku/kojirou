package api

import (
	"encoding/json"
	"fmt"
	"time"
)

type Localized map[string]string

type StringID = string

type Manga struct {
	Result        string
	Data          MangaData
	Relationships Relationships
}

type MangaData struct {
	ID         StringID
	Type       string
	Attributes struct {
		Title                  Localized
		AltTitles              []Localized
		Description            Localized
		IsLocked               bool
		Links                  map[string]string
		OriginalLanguage       string
		LastVolume             string
		LastChapter            string
		PublicationDemographic string
		Status                 string
		Year                   int
		ContentRating          string
		Tags                   Relationships
		CreatedAt              time.Time
		UpdatedAt              time.Time
		Version                int
	}
}

type Feed struct {
	Results []Chapter
	Limit   int
	Offset  int
	Total   int
}

type Chapter struct {
	Result        string
	Data          ChapterData
	Relationships Relationships
}

type ChapterData struct {
	ID         StringID
	Type       string
	Attributes struct {
		Volume             int
		Chapter            string
		Title              string
		TranslatedLanguage string
		Hash               string
		Data               []string
		DataSaver          []string
		PublishAt          time.Time
		CreatedAt          time.Time
		UpdatedAt          time.Time
		Version            int
	}
}

type Author struct {
	Result        string
	Data          AuthorData
	Relationships Relationships
}

type AuthorData struct {
	ID         StringID
	Type       string
	Attributes struct {
		Name      string
		ImageUrl  string
		Biography []string
		CreatedAt time.Time
		UpdatedAt time.Time
		Version   int
	}
}

type Group struct {
	Result        string
	Data          GroupData
	Relationships Relationships
}

type GroupData struct {
	ID         StringID
	Type       string
	Attributes struct {
		Name      string
		Leader    Relationship
		Members   Relationships
		CreatedAt time.Time
		UpdatedAt time.Time
		Version   int
	}
}

type AtHome struct {
	BaseURL string
}

type Relationships struct {
	Manga      []StringID
	Chapter    []StringID
	Author     []StringID
	Artist     []StringID
	Group      []StringID
	Tag        []StringID
	User       []StringID
	CustomList []StringID
}

func (rs *Relationships) UnmarshalJSON(data []byte) error {
	parsed := make([]Relationship, 0)
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	for _, r := range parsed {
		switch r.Type {
		case "manga":
			rs.Manga = append(rs.Manga, r.ID)
		case "chapter":
			rs.Chapter = append(rs.Chapter, r.ID)
		case "author":
			rs.Author = append(rs.Author, r.ID)
		case "artist":
			rs.Artist = append(rs.Artist, r.ID)
		case "scanlation_group":
			rs.Group = append(rs.Group, r.ID)
		case "tag":
			rs.Tag = append(rs.Tag, r.ID)
		case "user":
			rs.User = append(rs.User, r.ID)
		case "custom_list":
			rs.CustomList = append(rs.CustomList, r.ID)
		default:
			return fmt.Errorf("unsupported relationship: %v", r.Type)
		}
	}

	return nil
}

type Relationship struct {
	ID         StringID
	Type       string
	Attributes interface{}
}

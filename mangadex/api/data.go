package api

type Base struct {
	Code   int
	Status string
	Data   BaseData
}

type BaseData struct {
	ID           int
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
	ID       int
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
	ID         int
	Hash       string
	MangaID    int
	ThreadID   int
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
	ID   int
	Name string
}

type Chapter struct {
	Code   int
	Status string
	Data   ChapterData
}

type ChapterData struct {
	ID             int
	Hash           string
	MangaID        int
	ThreadID       int
	MangaTitle     string
	Volume         string
	Chapter        string
	Title          string
	Language       string
	Groups         []GroupMapping
	Uploader       int
	Timestamp      int
	Comments       int
	Views          int
	Status         string
	Pages          []string
	Server         string
	ServerFallback string
}

type Covers struct {
	Code   int
	Status string
	Data   CoversData
}

type CoversData []CoversMapping

type CoversMapping struct {
	Volume string
	URL    string
}

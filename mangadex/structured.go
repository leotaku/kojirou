package mangadex

import (
	"image"
	"sort"
)

type Manga struct {
	Info    MangaInfo
	Volumes map[Identifier]Volume
}

type Volume struct {
	Cover      image.Image
	Identifier Identifier
	Chapters   map[Identifier]Chapter
}

type Chapter struct {
	Info       ChapterInfo
	Identifier Identifier
	Pages      map[int]image.Image
}

func (m Manga) Sorted() []Volume {
	result := make([]Volume, 0)
	for _, idx := range m.Keys() {
		result = append(result, m.Volumes[idx])
	}

	return result
}

func (m Manga) Chapters() ChapterList {
	result := make(ChapterList, 0)
	for _, vol := range m.Volumes {
		for _, chap := range vol.Chapters {
			result = append(result, chap.Info)
		}
	}

	return result
}

func (m Manga) Keys() []Identifier {
	result := make([]Identifier, 0)
	for key := range m.Volumes {
		result = append(result, key)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Less(result[j])
	})

	return result
}

func (m Volume) Sorted() []Chapter {
	result := make([]Chapter, 0)
	for _, idx := range m.Keys() {
		result = append(result, m.Chapters[idx])
	}

	return result
}

func (m Volume) Keys() []Identifier {
	result := make([]Identifier, 0)
	for key := range m.Chapters {
		result = append(result, key)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Less(result[j])
	})

	return result
}

func (m Chapter) Sorted() []image.Image {
	result := make([]image.Image, len(m.Pages))
	for key, val := range m.Pages {
		result[key] = val
	}

	return result
}

func (m Manga) WithChapters(chapters ChapterList) Manga {
	m.Volumes = make(map[Identifier]Volume)
	for _, info := range chapters {
		chapID := info.Identifier
		volID := info.VolumeIdentifier
		if vol, ok := m.Volumes[volID]; ok {
			if _, ok := vol.Chapters[chapID]; !ok {
				m.Volumes[volID].Chapters[chapID] = extractChapter(info)
			}
		} else {
			m.Volumes[volID] = extractVolume(info)
		}
	}

	return m
}

func (m Manga) WithPages(pages ImageList) Manga {
	for idx, vol := range m.Volumes {
		vol.Chapters = make(map[Identifier]Chapter)
		m.Volumes[idx] = vol
	}
	for _, in := range pages {
		if _, ok := m.Volumes[in.volumeID].Chapters[in.chapterID]; ok {
			m.Volumes[in.volumeID].Chapters[in.chapterID].Pages[in.imageID] = in.Image
		}
	}

	return m
}

func (m Manga) WithCovers(covers ImageList) Manga {
	for idx, vol := range m.Volumes {
		vol.Cover = nil
		m.Volumes[idx] = vol
	}
	for _, in := range covers {
		if val, ok := m.Volumes[in.volumeID]; ok {
			val.Cover = in.Image
			m.Volumes[in.volumeID] = val
		}
	}

	return m
}

func extractVolume(info ChapterInfo) Volume {
	chapters := make(map[Identifier]Chapter)
	chapters[info.Identifier] = extractChapter(info)
	return Volume{
		Chapters:   chapters,
		Identifier: info.VolumeIdentifier,
	}
}

func extractChapter(info ChapterInfo) Chapter {
	return Chapter{
		Info:       info,
		Identifier: info.Identifier,
		Pages:      make(map[int]image.Image),
	}
}

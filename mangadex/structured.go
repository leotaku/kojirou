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

func (v Volume) Sorted() []Chapter {
	result := make([]Chapter, 0)
	for _, idx := range v.Keys() {
		result = append(result, v.Chapters[idx])
	}

	return result
}

func (v Volume) Keys() []Identifier {
	result := make([]Identifier, 0)
	for key := range v.Chapters {
		result = append(result, key)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Less(result[j])
	})

	return result
}

func (c Chapter) Sorted() []image.Image {
	result := make([]image.Image, len(c.Pages))
	for key, val := range c.Pages {
		result[key] = val
	}

	return result
}

func (m Manga) WithChapters(chapters ChapterList) Manga {
	vols := make(map[Identifier]Volume)
	for _, info := range chapters {
		chapID := info.Identifier
		volID := info.VolumeIdentifier
		if vol, ok := vols[volID]; ok {
			if _, ok := vol.Chapters[chapID]; !ok {
				vols[volID].Chapters[chapID] = extractChapter(info)
			}
		} else {
			vol := extractVolume(info)
			if val, ok := m.Volumes[volID]; ok {
				vol.Cover = val.Cover
			}
			vols[volID] = vol
		}
	}

	return Manga{
		Info:    m.Info,
		Volumes: vols,
	}
}

func (m Manga) WithPages(pages ImageList) Manga {
	vols := make(map[Identifier]Volume)
	for Idx, vol := range m.Volumes {
		vols[Idx] = vol
		for idx, chap := range vol.Chapters {
			chap.Pages = make(map[int]image.Image)
			m.Volumes[Idx].Chapters[idx] = chap
		}
	}
	for _, it := range pages {
		if _, ok := vols[it.volumeID].Chapters[it.chapterID]; ok {
			vols[it.volumeID].Chapters[it.chapterID].Pages[it.imageID] = it.Image
		}
	}

	return Manga{
		Info:    m.Info,
		Volumes: vols,
	}
}

func (m Manga) WithCovers(covers ImageList) Manga {
	vols := make(map[Identifier]Volume)
	for idx, vol := range m.Volumes {
		vol.Cover = nil
		vols[idx] = vol
	}
	for _, it := range covers {
		if vol, ok := vols[it.volumeID]; ok {
			vol.Cover = it.Image
			vols[it.volumeID] = vol
		}
	}

	return Manga{
		Info:    m.Info,
		Volumes: vols,
	}
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

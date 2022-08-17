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
	Info     VolumeInfo
	Chapters map[Identifier]Chapter
	Cover    image.Image
}

type Chapter struct {
	Info      ChapterInfo
	Pages     map[int]image.Image
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
			result = append(result, chap)
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

func (v Volume) Sorted() ChapterList {
	result := make(ChapterList, 0)
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

func (c Chapter) Keys() []int {
	result := make([]int, 0)
	for key := range c.Pages {
		result = append(result, key)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

func (c Chapter) Sorted() []image.Image {
	result := make([]image.Image, 0)
	for _, key := range c.Keys() {
		result = append(result, c.Pages[key])
	}

	return result
}

func (m Manga) WithChapters(chapters ChapterList) Manga {
	vols := make(map[Identifier]Volume)
	for _, chapter := range chapters {
		chapID := chapter.Info.Identifier
		volID := chapter.Info.VolumeIdentifier
		if vol, ok := vols[volID]; ok {
			if _, ok := vol.Chapters[chapID]; !ok {
				vols[volID].Chapters[chapID] = cleanChapter(chapter)
			}
		} else {
			vol := cleanVolume(chapter)
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
		if _, ok := vols[it.VolumeIdentifier].Chapters[it.ChapterIdentifier]; ok {
			vols[it.VolumeIdentifier].Chapters[it.ChapterIdentifier].Pages[it.ImageIdentifier] = it.Image
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
		if vol, ok := vols[it.VolumeIdentifier]; ok {
			vol.Cover = it.Image
			vols[it.VolumeIdentifier] = vol
		}
	}

	return Manga{
		Info:    m.Info,
		Volumes: vols,
	}
}

func cleanVolume(old Chapter) Volume {
	chapters := make(map[Identifier]Chapter)
	chapters[old.Info.Identifier] = cleanChapter(old)
	return Volume{
		Info: VolumeInfo{
			Identifier: old.Info.VolumeIdentifier,
		},
		Chapters: chapters,
	}
}

func cleanChapter(old Chapter) Chapter {
	pages := make(map[int]image.Image)
	for key, value := range old.Pages {
		pages[key] = value
	}

	return Chapter{
		Info:      old.Info,
		Pages:     pages,
	}
}

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
	Chapters   map[Identifier]Chapter
	Info     VolumeInfo
}

type Chapter struct {
	Info       ChapterInfo
	Identifier Identifier
	Pages      map[int]image.Image
	PagePaths []string
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
	return Chapter{
		Info:      old.Info,
		Pages:     make(map[int]image.Image),
		PagePaths: old.PagePaths,
	}
}

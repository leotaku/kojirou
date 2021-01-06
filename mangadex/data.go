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

func (m Manga) Sorted() []Identifier {
	result := make([]Identifier, 0)
	for key := range m.Volumes {
		result = append(result, key)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Less(result[j])
	})

	return result
}

func (m Volume) Sorted() []Identifier {
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

func Rebuild(mi MangaInfo, ci []ChapterInfo) Manga {
	vols := make(map[Identifier]Volume)
	for _, info := range ci {
		chapId := info.Identifier
		volId := info.VolumeIdentifier
		if vol, ok := vols[volId]; ok {
			if _, ok := vol.Chapters[chapId]; !ok {
				vols[volId].Chapters[chapId] = extractChapter(info)
			}
		} else {
			vols[volId] = extractVolume(info)
		}
	}

	return Manga{
		Info:    mi,
		Volumes: vols,
	}
}

func (m Manga) WithPages(pages []ImageInfo) Manga {
	for _, in := range pages {
		m.Volumes[in.volumeId].Chapters[in.chapterId].Pages[in.imageId] = in.Image
	}

	return m
}

func (m Manga) WithCovers(covers []ImageInfo) Manga {
	for _, in := range covers {
		val := m.Volumes[in.volumeId]
		val.Cover = in.Image
		m.Volumes[in.volumeId] = val
	}

	return m
}

func extractManga(mi MangaInfo, info ChapterInfo) Manga {
	volumes := make(map[Identifier]Volume)
	volumes[info.VolumeIdentifier] = extractVolume(info)

	return Manga{
		Volumes: volumes,
		Info:    mi,
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

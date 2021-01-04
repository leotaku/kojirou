package mangadex

type FullVersion struct {
	Manga *Manga
	*Version
}

func (m *Manga) Index(i int) *FullVersion {
	return &FullVersion{
		Manga:   m,
		Version: &m.Versions[i],
	}
}

type FullVolume struct {
	Manga   *Manga
	Version *Version
	*Volume
}

func (v *FullVersion) Index(i int) *FullVolume {
	return &FullVolume{
		Manga:   v.Manga,
		Version: v.Version,
		Volume:  &v.Volumes[i],
	}
}

type FullChapter struct {
	Manga   *Manga
	Version *Version
	Volume  *Volume
	*Chapter
}

func (v *FullVolume) Index(i int) *FullChapter {
	return &FullChapter{
		Manga:   v.Manga,
		Version: v.Version,
		Volume:  v.Volume,
		Chapter: &v.Chapters[i],
	}
}

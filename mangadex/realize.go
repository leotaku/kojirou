package mangadex

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/leotaku/manki/mangadex/api"
)

type HttpStatusError = api.HttpStatusError

func (f ChapterInfo) GetImagePaths() ([]PathInfo, error) {
	resp, err := api.FetchChapter(f.Id)
	if err != nil {
		return nil, err
	}

	result := make([]PathInfo, 0)
	for i, url := range chapterImages(resp.Data) {
		result = append(result, PathInfo{
			Url:       url,
			imageId:   i,
			chapterId: f.Identifier,
			volumeId:  f.VolumeIdentifier,
		})
	}
	return result, nil
}

func (i PathInfo) GetImage() (*ImageInfo, error) {
	img, err := fetchImage(i.Url)
	if err != nil {
		return nil, err
	}

	return &ImageInfo{
		Image:     img,
		chapterId: i.chapterId,
		volumeId:  i.volumeId,
		imageId:   i.imageId,
	}, nil
}

func chapterImages(c api.ChapterData) []string {
	result := make([]string, 0)
	for _, path := range c.Pages {
		url := fmt.Sprintf("%v/%v/%v", c.Server, c.Hash, path)
		result = append(result, url)
	}
	return result
}

func fetchImage(url string, a ...interface{}) (image.Image, error) {
	resp, err := http.Get(fmt.Sprintf(url, a...))
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	dec, _, err := image.Decode(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return dec, nil
}

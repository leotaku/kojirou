package mangadex

import (
	"errors"
	"fmt"
	"image"
	"net/http"

	"github.com/leotaku/manki/mangadex/api"
)

type HttpStatusError = api.HttpStatusError

func (c *Chapter) Realize() error {
	data, err := api.FetchChapter(c.Id)
	if err != nil {
		return err
	}
	c.Images = chapterImages(data.Data)

	return nil
}

func (i *Image) Realize() error {
	if len(i.Url) > 0 {
		img, err := fetchImage(i.Url)
		if err != nil {
			return err
		}
		i.Image = img
	}

	return nil
}

func chapterImages(c api.ChapterData) []Image {
	result := make([]Image, 0)
	for _, url := range c.Pages {
		result = append(result, Image{
			Url: fmt.Sprintf("%v/%v/%v", c.Server, c.Hash, url),
		})
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

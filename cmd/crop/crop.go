package crop

import (
	"fmt"
	"image"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func Crop(img image.Image, bounds image.Rectangle) (image.Image, error) {
	if img, ok := img.(SubImager); !ok {
		return nil, fmt.Errorf("image does not support cropping")
	} else {
		return img.SubImage(bounds), nil
	}
}

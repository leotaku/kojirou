package crop

import (
	"fmt"
	"image"
)

const aspectRatioLimit = 1.2

func ShouldSplit(img image.Image) bool {
	size := img.Bounds().Size()
	aspectRatio := float32(size.X) / float32(size.Y)

	return aspectRatio > aspectRatioLimit
}

func Split(img image.Image) (image.Image, image.Image, error) {
	bounds := img.Bounds()

	left := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+bounds.Dx()/2, bounds.Max.Y)
	right := image.Rect(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)

	if img, ok := img.(SubImager); !ok {
		return nil, nil, fmt.Errorf("image does not support cropping")
	} else {
		return img.SubImage(left), img.SubImage(right), nil
	}
}

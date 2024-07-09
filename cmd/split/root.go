package split

import (
	"fmt"
	"image"
)

func IsDoublePage(img image.Image) bool {
	bounds := img.Bounds()
	return bounds.Dx() >= bounds.Dy()
}

func GetNextImageIdentifier(current int, occupied map[int]bool) int {
	// while the current identifier is already occupied, increment it
	for occupied[current] {
		current++
	}
	return current
}

func SplitVertically(img image.Image) (image.Image, image.Image, error) {
	type subImager interface {
		image.Image
		SubImage(r image.Rectangle) image.Image
	}

	subImg, ok := img.(subImager)
	if !ok {
		return nil, nil, fmt.Errorf("image does not support splitting or not a valid image")
	}

	originalBounds := subImg.Bounds()
	xMiddle := originalBounds.Dx() / 2

	leftBounds := image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{xMiddle, originalBounds.Dy()},
	}
	rightBounds := image.Rectangle{
		Min: image.Point{xMiddle, 0},
		Max: image.Point{originalBounds.Dx(), originalBounds.Dy()},
	}

	leftImage := subImg.SubImage(leftBounds)
	rightImage := subImg.SubImage(rightBounds)

	return leftImage, rightImage, nil
}

func RotateImage(img image.Image) (image.Image, error) {
	originalBounds := img.Bounds()
	width := originalBounds.Dx()
	height := originalBounds.Dy()

	// create a new empty image with rotated dimensions
	rotatedImage := image.NewRGBA(image.Rect(0, 0, height, width))

	// rotate the image by mapping each pixel from the original to the new position
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotatedImage.Set(y, width-1-x, img.At(x, y))
		}
	}

	return rotatedImage, nil
}
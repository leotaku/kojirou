package crop

import (
	"fmt"
	"image"
	"image/color"
)

const grayDarknessLimit = 128

func Crop(img image.Image, bounds image.Rectangle) (image.Image, error) {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	if simg, ok := img.(subImager); !ok {
		return nil, fmt.Errorf("image does not support cropping")
	} else {
		return simg.SubImage(bounds), nil
	}
}

func Limited(img image.Image, limit float32) image.Rectangle {
	bounds := img.Bounds()
	maxPixels := float32((bounds.Dx() + bounds.Dy()) / 2) * limit
	return Bounds(img).Union(bounds.Inset(int(maxPixels)))
}

func Bounds(img image.Image) image.Rectangle {
	left := findBorder(img, image.Pt(1, 0))
	right := findBorder(img, image.Pt(-1, 0))
	top := findBorder(img, image.Pt(0, 1))
	bottom := findBorder(img, image.Pt(0, -1))

	return image.Rect(left.X, top.Y, right.X, bottom.Y)
}

func findBorder(img image.Image, dir image.Point) image.Point {
	bounds := img.Bounds()
	scan := image.Pt(dir.Y, dir.X)
	dpt := pointInScanCorner(bounds, dir)

	for !scanLineForNonWhitespace(img, dpt, scan) {
		dpt = dpt.Add(dir)
		if !dpt.In(bounds) {
			dpt = pointInScanCorner(bounds, dir)
			break
		}
	}

	if dir.X < 0 || dir.Y < 0 {
		return dpt.Sub(dir)
	} else {
		return dpt
	}
}

func pointInScanCorner(rect image.Rectangle, dir image.Point) image.Point {
	if dir.X < 0 || dir.Y < 0 {
		return rect.Max.Sub(image.Pt(1, 1))
	} else {
		return rect.Min
	}
}

func scanLineForNonWhitespace(img image.Image, pt image.Point, scan image.Point) bool {
	for spt := pt; spt.In(img.Bounds()); spt = spt.Add(scan) {
		if gray, ok := color.GrayModel.Convert(img.At(spt.X, spt.Y)).(color.Gray); ok {
			if gray.Y <= grayDarknessLimit {
				return true
			}
		}
	}

	return false
}

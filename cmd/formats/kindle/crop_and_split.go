package kindle

import (
	"image"

	"github.com/leotaku/kojirou/cmd/crop"
)

type WidepagePolicy int

const (
	WidepagePolicyPreserve WidepagePolicy = iota
	WidepagePolicySplit
	WidepagePolicyPreserveAndSplit
	WidepagePolicySplitAndPreserve
)

func cropAndSplit(img image.Image, widepage WidepagePolicy, autocrop bool, ltr bool) []image.Image {
	if autocrop {
		croppedImg, err := crop.Crop(img, crop.Bounds(img))
		if err != nil {
			panic("unsupported image type for splitting")
		}
		img = croppedImg
	}

	if widepage != WidepagePolicyPreserve && crop.ShouldSplit(img) {
		left, right, err := crop.Split(img)
		if err != nil {
			panic("unsupported image type for splitting")
		}

		switch widepage {
		case WidepagePolicySplit:
			if ltr {
				return []image.Image{left, right}
			} else {
				return []image.Image{right, left}
			}
		case WidepagePolicyPreserveAndSplit:
			if ltr {
				return []image.Image{img, left, right}
			} else {
				return []image.Image{img, right, left}
			}
		case WidepagePolicySplitAndPreserve:
			if ltr {
				return []image.Image{left, right, img}
			} else {
				return []image.Image{right, left, img}
			}
		}
	}

	return []image.Image{img}
}

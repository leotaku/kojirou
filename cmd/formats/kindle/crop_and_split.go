package kindle

import (
	"image"

	"github.com/leotaku/kojirou/cmd/crop"
)

type AutosplitPolicy int

const (
	AutosplitPolicyPreserve AutosplitPolicy = iota
	AutosplitPolicySplit
	AutosplitPolicyBoth
)

func cropAndSplit(img image.Image, autosplit AutosplitPolicy, autocrop bool, ltr bool) []image.Image {
	if autocrop {
		croppedImg, err := crop.Crop(img, crop.Bounds(img))
		if err != nil {
			panic("unsupported image type for splitting")
		}
		img = croppedImg
	}

	if autosplit != AutosplitPolicyPreserve && crop.ShouldSplit(img) {
		left, right, err := crop.Split(img)
		if err != nil {
			panic("unsupported image type for splitting")
		}

		switch autosplit {
		case AutosplitPolicySplit:
			if ltr {
				return []image.Image{left, right}
			} else {
				return []image.Image{right, left}
			}
		case AutosplitPolicyBoth:
			if ltr {
				return []image.Image{left, right, img}
			} else {
				return []image.Image{right, left, img}
			}
		}
	}

	return []image.Image{img}
}

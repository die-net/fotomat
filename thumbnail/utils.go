package thumbnail

import (
	"github.com/die-net/fotomat/vips"
)

// Scale original (width, height) to result (width, height), maintaining aspect ratio.
// If within=true, fit completely within result, leaving empty space if necessary.
func scaleAspect(ow, oh, rw, rh int, within bool) (int, int, bool) {
	// Scale aspect ratio using integer math, avoiding floating point
	// errors.

	wp := ow * rh
	hp := oh * rw

	trustWidth := false
	if within == (wp < hp) {
		rw = (wp + oh - 1) / oh
	} else {
		rh = (hp + ow - 1) / ow
		trustWidth = true
	}

	if rw < 1 {
		rw = 1
	}
	if rh < 1 {
		rh = 1
	}

	return rw, rh, trustWidth
}

func preShrinkFactor(mw, mh, iw, ih int, trustWidth bool, jpeg bool) int {
	// JPEG shrink on VIPS >= 8.6.4 and WebP shrink both round down the
	// number of pixels.  Round our shrink factor down by a pixel to
	// make sure we are never left with too few.
	var shrink float64
	if trustWidth {
		shrink = float64(mw) / (float64(iw) * fastResizeLimit)
		shrink -= (shrink - 1) / float64(iw)
	} else {
		shrink = float64(mh) / (float64(ih) * fastResizeLimit)
		shrink -= (shrink - 1) / float64(ih)
	}

	// Jpeg loader can quickly shrink by 2, 4, or 8.
	if jpeg {
		switch {
		case shrink >= 8:
			return 8
		case shrink >= 4:
			return 4
		case shrink >= 2:
			return 2
		default:
			return 1
		}
	}

	switch {
	case shrink >= 1024:
		return 1024
	case shrink >= 2:
		return int(shrink)
	default:
		return 1
	}
}

func minTransparency(image *vips.Image) (float64, error) {
	if !image.HasAlpha() {
		return 1.0, nil
	}

	band, err := image.Copy()
	if err != nil {
		return 0, err
	}
	defer band.Close()

	if err := band.ExtractBand(band.ImageGetBands()-1, 1); err != nil {
		return 0, err
	}

	// If all pixels are at least 90% opaque, we can flatten.
	min, err := band.Min()
	if err != nil {
		return 0, err
	}

	return min / band.MaxAlpha(), nil
}

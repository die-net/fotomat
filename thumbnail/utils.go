package thumbnail

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

func jpegShrink(mw, mh, iw, ih int, trustWidth bool) int {
	var shrink int
	if trustWidth {
		shrink = mw / iw
	} else {
		shrink = mh / ih
	}

	// Jpeg loader can quickly shrink by 2, 4, or 8.
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

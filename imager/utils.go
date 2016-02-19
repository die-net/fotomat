// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

// Scale original (width, height) to result (width, height), maintaining aspect ratio.
// If within=true, fit completely within result, leaving empty space if necessary.
func scaleAspect(ow, oh, rw, rh int, within bool) (int, int) {
	// Scale aspect ratio using integer math, avoiding floating point
	// errors.

	wp := ow * rh
	hp := oh * rw

	if within == (wp < hp) {
		rw = (wp + oh/2) / oh
	} else {
		rh = (hp + ow/2) / ow
	}

	return rw, rh
}

func jpegShrink(shrink int) int {
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

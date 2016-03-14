// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"gopkg.in/gographics/imagick.v2/imagick"
)

type Orientation struct {
	swapXY bool
	flipX  bool
	flipY  bool
	fn     func(*imagick.MagickWand) error
}

func NewOrientation(orientation imagick.OrientationType) *Orientation {
	switch orientation {
	default:
		return &Orientation{swapXY: false, flipX: false, flipY: false,
			fn: nil}
	case imagick.ORIENTATION_TOP_RIGHT:
		return &Orientation{swapXY: false, flipX: true, flipY: false,
			fn: func(wand *imagick.MagickWand) error { return wand.FlopImage() }}
	case imagick.ORIENTATION_BOTTOM_RIGHT:
		return &Orientation{swapXY: false, flipX: true, flipY: true,
			fn: func(wand *imagick.MagickWand) error { return wand.RotateImage(white, 180.0) }}
	case imagick.ORIENTATION_BOTTOM_LEFT:
		return &Orientation{swapXY: false, flipX: false, flipY: true,
			fn: func(wand *imagick.MagickWand) error { return wand.FlipImage() }}
	case imagick.ORIENTATION_LEFT_TOP:
		return &Orientation{swapXY: true, flipX: false, flipY: false,
			fn: func(wand *imagick.MagickWand) error { return wand.TransposeImage() }}
	case imagick.ORIENTATION_RIGHT_TOP:
		return &Orientation{swapXY: true, flipX: false, flipY: true,
			fn: func(wand *imagick.MagickWand) error { return wand.RotateImage(white, 90.0) }}
	case imagick.ORIENTATION_RIGHT_BOTTOM:
		return &Orientation{swapXY: true, flipX: true, flipY: true,
			fn: func(wand *imagick.MagickWand) error { return wand.TransverseImage() }}
	case imagick.ORIENTATION_LEFT_BOTTOM:
		return &Orientation{swapXY: true, flipX: true, flipY: false,
			fn: func(wand *imagick.MagickWand) error { return wand.RotateImage(white, 270.0) }}
	}
}

func (orientation *Orientation) Dimensions(width, height uint) (uint, uint) {
	if orientation.swapXY {
		return height, width
	}
	return width, height
}

func (orientation *Orientation) Crop(ow, oh uint, x, y int, iw, ih uint) (uint, uint, int, int) {
	if orientation.swapXY {
		ow, oh = oh, ow
		x, y = y, x
		iw, ih = ih, iw
	}
	if orientation.flipX {
		x = int(iw) - int(ow) - x
	}
	if orientation.flipY {
		y = int(ih) - int(oh) - y
	}
	return ow, oh, x, y
}

func (orientation *Orientation) Fix(wand *imagick.MagickWand) error {
	if orientation.fn == nil {
		return nil
	}
	if err := orientation.fn(wand); err != nil {
		return err
	}
	if err := wand.SetImageOrientation(imagick.ORIENTATION_TOP_LEFT); err != nil {
		return err
	}

	*orientation = Orientation{swapXY: false, flipX: false, flipY: false, fn: nil}
	return nil
}

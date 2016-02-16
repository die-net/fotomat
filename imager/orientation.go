// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"github.com/die-net/fotomat/vips"
	"strconv"
)

type Orientation int

const (
	Unknown Orientation = iota
	TopLeft
	TopRight
	BottomRight
	BottomLeft
	LeftTop
	RightTop
	RightBottom
	LeftBottom
)

var orientationInfo = []struct {
	swapXY bool
	flipX  bool
	flipY  bool
	apply  func(*vips.Image) (*vips.Image, error)
}{
	{swapXY: false, flipX: false, flipY: false, apply: nil}, // Unknown
	{swapXY: false, flipX: false, flipY: false, apply: nil},
	{swapXY: false, flipX: true, flipY: false, apply: func(image *vips.Image) (*vips.Image, error) { return image.Flip(vips.DirectionHorizontal) }},
	{swapXY: false, flipX: true, flipY: true, apply: func(image *vips.Image) (*vips.Image, error) { return image.Rot(vips.AngleD180) }},
	{swapXY: false, flipX: false, flipY: true, apply: func(image *vips.Image) (*vips.Image, error) { return image.Flip(vips.DirectionVertical) }},
	{swapXY: true, flipX: false, flipY: false, apply: func(image *vips.Image) (*vips.Image, error) { return image.Flip(vips.DirectionVertical) }}, //TODO: Transpose
	{swapXY: true, flipX: false, flipY: true, apply: func(image *vips.Image) (*vips.Image, error) { return image.Rot(vips.AngleD90) }},
	{swapXY: true, flipX: true, flipY: true, apply: func(image *vips.Image) (*vips.Image, error) { return image.Flip(vips.DirectionVertical) }}, //TODO: Transverse
	{swapXY: true, flipX: true, flipY: false, apply: func(image *vips.Image) (*vips.Image, error) { return image.Rot(vips.AngleD270) }},
}

func DetectOrientation(image *vips.Image) Orientation {
	o, ok := image.ImageGetAsString(vips.ExifOrientation)
	if !ok || o == "" {
		return Unknown
	}

	orientation, err := strconv.Atoi(o)
	if err != nil || orientation <= 0 || orientation >= len(orientationInfo) {
		return Unknown
	}

	return Orientation(orientation)
}

func (orientation Orientation) Dimensions(width, height int) (int, int) {
	if orientationInfo[orientation].swapXY {
		return height, width
	}
	return width, height
}

func (orientation Orientation) Crop(ow, oh int, x, y int, iw, ih int) (int, int, int, int) {
	oi := &orientationInfo[orientation]

	if oi.swapXY {
		ow, oh = oh, ow
		x, y = y, x
		iw, ih = ih, iw
	}
	if oi.flipX {
		x = int(iw) - int(ow) - x
	}
	if oi.flipY {
		y = int(ih) - int(oh) - y
	}
	return ow, oh, x, y
}

func (orientation Orientation) Apply(image *vips.Image) (*vips.Image, error) {
	oi := &orientationInfo[orientation]

	if oi.apply == nil {
		return image, nil
	}
	out, err := oi.apply(image)
	if err != nil {
		return nil, err
	}

	_ = out.ImageRemove(vips.ExifOrientation)

	return out, nil
}

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
	{swapXY: false, flipX: true, flipY: false, apply: flop},
	{swapXY: false, flipX: true, flipY: true, apply: rot180},
	{swapXY: false, flipX: false, flipY: true, apply: flip},
	{swapXY: true, flipX: false, flipY: false, apply: transpose},
	{swapXY: true, flipX: false, flipY: true, apply: rot90},
	{swapXY: true, flipX: true, flipY: true, apply: transverse},
	{swapXY: true, flipX: true, flipY: false, apply: rot270},
}

func DetectOrientation(image *vips.Image) Orientation {
	o, ok := image.ImageGetAsString(vips.ExifOrientation)
	if !ok || o == "" {
		return Unknown
	}

	orientation, err := strconv.Atoi(o[:1])
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
	return x, y, ow, oh
}

func (orientation Orientation) Apply(image *vips.Image) (*vips.Image, error) {
	oi := &orientationInfo[orientation]

	if oi.apply == nil {
		return nil, nil
	}

	// We want to stay sequential, so we copy memory here and execute
	// all work in the pipeline so far.
	out, err := image.Write()
	if err != nil {
		return nil, err
	}

	rot, err := oi.apply(out)
	out.Close()
	if err != nil {
		return nil, err
	}

	_ = rot.ImageRemove(vips.ExifOrientation)

	return rot, nil
}

func flip(image *vips.Image) (*vips.Image, error) {
	return image.Flip(vips.DirectionVertical)
}

func flop(image *vips.Image) (*vips.Image, error) {
	return image.Flip(vips.DirectionHorizontal)
}

func rot90(image *vips.Image) (*vips.Image, error) {
	return image.Rot(vips.Angle90)
}

func rot180(image *vips.Image) (*vips.Image, error) {
	return image.Rot(vips.Angle180)
}

func rot270(image *vips.Image) (*vips.Image, error) {
	return image.Rot(vips.Angle270)
}

func transpose(image *vips.Image) (*vips.Image, error) {
	flip, err := image.Flip(vips.DirectionVertical)
	if err != nil {
		return nil, err
	}
	defer flip.Close()
	return flip.Rot(vips.Angle90)
}

func transverse(image *vips.Image) (*vips.Image, error) {
	flip, err := image.Flip(vips.DirectionVertical)
	if err != nil {
		return nil, err
	}
	defer flip.Close()
	return flip.Rot(vips.Angle270)
}

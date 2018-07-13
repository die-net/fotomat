package format

import (
	"github.com/die-net/fotomat/vips"
	"strconv"
)

// Orientation is the current Image orientation, as stored by a camera.
type Orientation int

// Orientation values as defined by EXIF.
const (
	Undefined Orientation = iota
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
	apply  func(*vips.Image) error
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

// DetectOrientation detects the current Image Orientation from the EXIF header.
func DetectOrientation(image *vips.Image) Orientation {
	o, ok := image.ImageGetAsString(vips.ExifOrientation)
	if !ok || o == "" {
		return Undefined
	}

	orientation, err := strconv.Atoi(o[:1])
	if err != nil || orientation <= 0 || orientation >= len(orientationInfo) {
		return Undefined
	}

	return Orientation(orientation)
}

// Dimensions translates a virtual width and height to match the current physical Orientation.
func (orientation Orientation) Dimensions(width, height int) (int, int) {
	if orientationInfo[orientation].swapXY {
		return height, width
	}
	return width, height
}

// Crop translates crop parameters from virtual coordinates to match the current physical Orientation.
func (orientation Orientation) Crop(ow, oh, x, y, iw, ih int) (int, int, int, int) {
	oi := &orientationInfo[orientation]

	if oi.swapXY {
		ow, oh = oh, ow
		x, y = y, x
		iw, ih = ih, iw
	}
	if oi.flipX {
		x = iw - ow - x
	}
	if oi.flipY {
		y = ih - oh - y
	}
	return x, y, ow, oh
}

// Apply executes a set of operations to change the pixel ordering from
// orientation to TopLeft.
func (orientation Orientation) Apply(image *vips.Image) error {
	oi := &orientationInfo[orientation]

	if oi.apply == nil {
		return nil
	}

	// We want to stay sequential, so we copy memory here and execute
	// all work in the pipeline so far.
	if err := image.Write(); err != nil {
		return err
	}

	if err := oi.apply(image); err != nil {
		return err
	}

	_ = image.ImageRemove(vips.ExifOrientation)

	return nil
}

func flip(image *vips.Image) error {
	return image.Flip(vips.DirectionVertical)
}

func flop(image *vips.Image) error {
	return image.Flip(vips.DirectionHorizontal)
}

func rot90(image *vips.Image) error {
	return image.Rot(vips.Angle90)
}

func rot180(image *vips.Image) error {
	return image.Rot(vips.Angle180)
}

func rot270(image *vips.Image) error {
	return image.Rot(vips.Angle270)
}

func transpose(image *vips.Image) error {
	if err := image.Flip(vips.DirectionVertical); err != nil {
		return err
	}
	return image.Rot(vips.Angle90)
}

func transverse(image *vips.Image) error {
	if err := image.Flip(vips.DirectionVertical); err != nil {
		return err
	}
	return image.Rot(vips.Angle270)
}

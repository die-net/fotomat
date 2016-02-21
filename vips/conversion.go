package vips

/*
#cgo pkg-config: vips
#include "conversion.h"
*/
import "C"

type Extend int

const (
	ExtendBlack      Extend = C.VIPS_EXTEND_BLACK
	ExtendCopy       Extend = C.VIPS_EXTEND_COPY
	ExtendRepeat     Extend = C.VIPS_EXTEND_REPEAT
	ExtendMirror     Extend = C.VIPS_EXTEND_MIRROR
	ExtendWhite      Extend = C.VIPS_EXTEND_WHITE
	ExtendBackground Extend = C.VIPS_EXTEND_BACKGROUND
)

type Angle int

const (
	Angle0   Angle = C.VIPS_ANGLE_D0
	Angle90  Angle = C.VIPS_ANGLE_D90
	Angle180 Angle = C.VIPS_ANGLE_D180
	Angle270 Angle = C.VIPS_ANGLE_D270
)

type Direction int

const (
	DirectionHorizontal Direction = C.VIPS_DIRECTION_HORIZONTAL
	DirectionVertical   Direction = C.VIPS_DIRECTION_VERTICAL
)

func (in Image) Cast(format BandFormat) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_cast(in.vi, &out, C.VipsBandFormat(format))
	return imageError(out, e)
}

func (in Image) Copy() (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_copy(in.vi, &out)
	return imageError(out, e)
}

func (in Image) Embed(left, top, width, height int, extend Extend) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_embed(in.vi, &out, C.int(left), C.int(top), C.int(width), C.int(height), C.int(extend))
	return imageError(out, e)
}

func (in Image) ExtractArea(left, top, width, height int) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_extract_area(in.vi, &out, C.int(left), C.int(top), C.int(width), C.int(height))
	return imageError(out, e)
}

func (in Image) Flip(direction Direction) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_flip(in.vi, &out, C.VipsDirection(direction))
	return imageError(out, e)
}

func (in Image) Premultiply() (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_premultiply(in.vi, &out)
	return imageError(out, e)
}

func (in Image) Rot(angle Angle) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_rot(in.vi, &out, C.VipsAngle(angle))
	return imageError(out, e)
}

func (in Image) Unpremultiply() (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_unpremultiply(in.vi, &out)
	return imageError(out, e)
}

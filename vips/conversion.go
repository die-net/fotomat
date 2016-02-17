package vips

/*
#cgo pkg-config: vips
#include "conversion.h"
*/
import "C"

type Extend int

const (
	ExtendBlack      = C.VIPS_EXTEND_BLACK
	ExtendCopy       = C.VIPS_EXTEND_COPY
	ExtendRepeat     = C.VIPS_EXTEND_REPEAT
	ExtendMirror     = C.VIPS_EXTEND_MIRROR
	ExtendWhite      = C.VIPS_EXTEND_WHITE
	ExtendBackground = C.VIPS_EXTEND_BACKGROUND
)

type Angle int

const (
	Angle0   = C.VIPS_ANGLE_0
	Angle90  = C.VIPS_ANGLE_90
	Angle180 = C.VIPS_ANGLE_180
	Angle270 = C.VIPS_ANGLE_270
)

type Direction int

const (
	DirectionHorizontal = C.VIPS_DIRECTION_HORIZONTAL
	DirectionVertical   = C.VIPS_DIRECTION_VERTICAL
)

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

func (in Image) Rot(angle Angle) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_rot(in.vi, &out, C.VipsAngle(angle))
	return imageError(out, e)
}

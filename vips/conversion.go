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

func (in Image) Copy() (Image, error) {
	out := Image{}
	err := vipsError(C.cgo_vips_copy(in.image, &out.image))
	return out, err
}

func (in Image) Embed(left, top, width, height int, extend Extend) (Image, error) {
	out := Image{}
	err := vipsError(C.cgo_vips_embed(in.image, &out.image, C.int(left), C.int(top), C.int(width), C.int(height), C.int(extend)))
	return out, err
}

func (in Image) ExtractArea(left, top, width, height int) (Image, error) {
	out := Image{}
	err := vipsError(C.cgo_vips_extract_area(in.image, &out.image, C.int(left), C.int(top), C.int(width), C.int(height)))
	return out, err
}

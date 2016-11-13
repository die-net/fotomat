package vips

/*
#cgo pkg-config: vips
#include "conversion.h"
*/
import "C"

// Extend specifies how to extend edges of an image
type Extend int

// Various Extend values understand by VIPS.
const (
	ExtendBlack      Extend = C.VIPS_EXTEND_BLACK      // extends with black (all 0) pixels
	ExtendCopy       Extend = C.VIPS_EXTEND_COPY       // copies the image edges
	ExtendRepeat     Extend = C.VIPS_EXTEND_REPEAT     // repeats the whole image
	ExtendMirror     Extend = C.VIPS_EXTEND_MIRROR     // mirrors the whole image
	ExtendWhite      Extend = C.VIPS_EXTEND_WHITE      // extends with white (all bits set) pixels
	ExtendBackground Extend = C.VIPS_EXTEND_BACKGROUND // extends with colour from the background property
)

// Angle specifies fixed rotation angles
type Angle int

// Various Angle values understood by VIPS.
const (
	Angle0   Angle = C.VIPS_ANGLE_D0   // does not rotate
	Angle90  Angle = C.VIPS_ANGLE_D90  // 90 degrees counter-clockwise
	Angle180 Angle = C.VIPS_ANGLE_D180 // 180 degrees
	Angle270 Angle = C.VIPS_ANGLE_D270 // 90 degrees clockwise
)

// Direction specifies which direction to flip an image
type Direction int

// Various Direction values undertood by VIPS.
const (
	DirectionHorizontal Direction = C.VIPS_DIRECTION_HORIZONTAL // left-right
	DirectionVertical   Direction = C.VIPS_DIRECTION_VERTICAL   // top-bottom
)

// Cast converts in to BandFormat. Floats are truncated (not rounded). Out of range values are clipped.
func (in *Image) Cast(format BandFormat) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_cast(in.vi, &out, C.VipsBandFormat(format))
	return in.imageError(out, e)
}

// Copy an image by copying pointers, so this operation is instant, even for very large images.
func (in *Image) Copy() (*Image, error) {
	var out *C.struct__VipsImage
	err := vipsError(C.cgo_vips_copy(in.vi, &out))
	return imageFromVi(out), err
}

// Embed in within an image of size width by height at position x, y.
// Extend controls what appears in the new pixels.
func (in *Image) Embed(left, top, width, height int, extend Extend) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_embed(in.vi, &out, C.int(left), C.int(top), C.int(width), C.int(height), C.int(extend))
	return in.imageError(out, e)
}

// ExtractArea extract an area from an image. The area must fit within in.
func (in *Image) ExtractArea(left, top, width, height int) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_extract_area(in.vi, &out, C.int(left), C.int(top), C.int(width), C.int(height))
	return in.imageError(out, e)
}

// ExtractBand extracts band (channel) number n from in.  Extracting out of range is an error.
func (in *Image) ExtractBand(band, n int) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_extract_band(in.vi, &out, C.int(band), C.int(n))
	return in.imageError(out, e)
}

// Flatten takes the last band of in as an alpha and use it to blend the remaining
// channels with black, then remove the alpha channel.
func (in *Image) Flatten() error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_flatten(in.vi, &out)
	return in.imageError(out, e)
}

// Flip an image left-right or up-down.
func (in *Image) Flip(direction Direction) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_flip(in.vi, &out, C.VipsDirection(direction))
	return in.imageError(out, e)
}

// MaxAlpha returns the maximum value for an alpha channel in current BandFormat of image.
func (in *Image) MaxAlpha() float64 {
	return float64(C.cgo_max_alpha(in.vi))
}

// Premultiply any alpha channel. The final band is taken to be the alpha.
func (in *Image) Premultiply() error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_premultiply(in.vi, &out)
	return in.imageError(out, e)
}

// Rot rotates an image by a fixed angle.
func (in *Image) Rot(angle Angle) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_rot(in.vi, &out, C.VipsAngle(angle))
	return in.imageError(out, e)
}

// Unpremultiply any alpha channel. The final band is taken to be the alpha.
func (in *Image) Unpremultiply() error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_unpremultiply(in.vi, &out)
	return in.imageError(out, e)
}

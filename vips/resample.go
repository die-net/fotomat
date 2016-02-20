package vips

/*
#cgo pkg-config: vips
#include "resample.h"
*/
import "C"

// Interpolate is an instance of an interpolator used by Affine.
type Interpolate struct {
	interpolate *C.struct__VipsInterpolate
}

// NewInterpolate creates an Interpolate instance from a nickname (nearest,
// bilinear, bicubic, lbb, nohalo, or vsqbs) and makes one.  Must be closed
// when you're done with it.
func NewInterpolate(name string) *Interpolate {
	interpolate := C.vips_interpolate_new(C.CString(name))
	if interpolate == nil {
		return nil
	}
	return &Interpolate{interpolate: interpolate}
}

// Close frees resources from an Interpolate.
func (i Interpolate) Close() {
	C.g_object_unref(C.gpointer(i.interpolate))
	i.interpolate = nil
}

// Affine performs this affine transform on the input image using the
// supplied interpolate:
//   X = a * (x + idx ) + b * (y + idy ) + odx
//   Y = c * (x + idx ) + d * (y + idy ) + doy
// x and y are the coordinates in input image. X and Y are the coordinates
// in output image.  (0,0) is the upper left corner.
func (in *Image) Affine(a, b, c, d float64, interpolate *Interpolate) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_affine(in.vi, &out, C.double(a), C.double(b), C.double(c), C.double(d), interpolate.interpolate)
	return imageError(out, e)
}

// Resize an image using the bicubic interpolator. When upsizing (scale >
// 1), the image is simply resized with Affine().  When downsizing, the
// image is block-shrunk with Shrink() to roughly half the interpolator
// window size above the target size, then blurred with an anti-alias
// filter, then resampled with Affine(), then sharpened.
func (in *Image) Resize(scale float64) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_resize(in.vi, &out, C.double(scale))
	return imageError(out, e)
}

// Shrink in by a pair of factors with a simple box filter.  You will get
// aliasing for non-integer shrinks.  In this case, shrink with this
// function to the nearest integer size above the target shrink, then
// downsample to the exact size with Affine().
func (in *Image) Shrink(xshrink, yshrink float64) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_shrink(in.vi, &out, C.double(xshrink), C.double(yshrink))
	return imageError(out, e)
}

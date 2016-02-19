package vips

/*
#cgo pkg-config: vips
#include "convolution.h"
*/
import "C"

func (in Image) Gaussblur(sigma float64) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_gaussblur(in.vi, &out, C.double(sigma))
	return imageError(out, e)
}

func (in Image) Sharpen(radius int, m1, m2 float64) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_sharpen(in.vi, &out, C.int(radius), C.double(m1), C.double(m2))
	return imageError(out, e)
}

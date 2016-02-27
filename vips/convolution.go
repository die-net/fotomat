package vips

/*
#cgo pkg-config: vips
#include "convolution.h"
*/
import "C"

func (in *Image) Gaussblur(sigma float64) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_gaussblur(in.vi, &out, C.double(sigma))
	return in.imageError(out, e)
}

func (in *Image) MildSharpen() error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_mild_sharpen(in.vi, &out)
	return in.imageError(out, e)
}

func (in *Image) Sharpen(radius int, x1, y2, y3, m1, m2 float64) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_sharpen(in.vi, &out, C.int(radius), C.double(x1), C.double(y2), C.double(y3), C.double(m1), C.double(m2))
	return in.imageError(out, e)
}

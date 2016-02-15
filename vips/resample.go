package vips

/*
#cgo pkg-config: vips
#include "resample.h"
*/
import "C"

type Interpolate struct {
	interpolate *C.struct__VipsInterpolate
}

func NewInterpolate(name string) *Interpolate {
	interpolate := C.vips_interpolate_new(C.CString(name))
	if interpolate == nil {
		return nil
	}
	return &Interpolate{interpolate: interpolate}
}

func (i Interpolate) Close() {
	C.g_object_unref(C.gpointer(i.interpolate))
	i.interpolate = nil
}

func (in *Image) Affine(a, b, c, d float64, interpolate Interpolate) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_affine(in.vi, &out, C.double(a), C.double(b), C.double(c), C.double(d), interpolate.interpolate)
	return imageError(out, e)
}

func (in *Image) Resize(scale float64, interpolate Interpolate) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_resize(in.vi, &out, C.double(scale), interpolate.interpolate)
	return imageError(out, e)
}

func (in *Image) Shrink(xshrink, yshrink float64) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_shrink(in.vi, &out, C.double(xshrink), C.double(yshrink))
	return imageError(out, e)
}

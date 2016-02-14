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

func (in *VipsImage) Affine(a, b, c, d float64, interpolate Interpolate) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_affine(in.image, &out.image, C.double(a), C.double(b), C.double(c), C.double(d), interpolate.interpolate))
	return out, err
}

func (in *VipsImage) Resize(scale float64, interpolate Interpolate) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_resize(in.image, &out.image, C.double(scale), interpolate.interpolate))
	return out, err
}

func (in *VipsImage) Shrink(xshrink, yshrink float64) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_shrink(in.image, &out.image, C.double(xshrink), C.double(yshrink)))
	return out, err
}

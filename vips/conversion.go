package vips

/*
#cgo pkg-config: vips
#include "conversion.h"
*/
import "C"

func (in VipsImage) Copy() (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_copy(in.image, &out.image))
	return out, err
}

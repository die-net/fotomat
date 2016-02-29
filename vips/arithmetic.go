package vips

/*
#cgo pkg-config: vips
#include "arithmetic.h"
*/
import "C"

func (in *Image) Min() (float64, error) {
	var out C.double
	err := vipsError(C.cgo_vips_min(in.vi, &out))
	return float64(out), err
}

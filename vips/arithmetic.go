package vips

/*
#cgo pkg-config: vips
#include "arithmetic.h"
*/
import "C"

// Min finds the single smallest value in all bands of the input image.
func (in *Image) Min() (float64, error) {
	var out C.double
	err := vipsError(C.cgo_vips_min(in.vi, &out))
	return float64(out), err
}

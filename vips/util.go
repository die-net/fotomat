package vips

/*
#cgo pkg-config: vips
#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>
*/
import "C"

import (
	"errors"
)

func btoi(b bool) int {
	if b {
		return 1
	}

	return 0
}

// vipsError() converts from vips to Go errors.
func vipsError(e C.int) error {
	if e == 0 {
		return nil
	}

	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()
	return errors.New(s)
}

// imageError() is a convenience wrapper around vipsError() for funcs that
// call a vips function passing in an output C.struct__VipsImage and return
// (Image, error).
func imageError(out *C.struct__VipsImage, e C.int) (*Image, error) {
	if err := vipsError(e); err != nil {
		return nil, err
	}

	return imageFromVi(out), nil
}

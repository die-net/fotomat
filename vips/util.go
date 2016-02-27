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

// btoi converts from Go boolean to int with value 0 or 1.
func btoi(b bool) int {
	if b {
		return 1
	}

	return 0
}

// vipsError converts from vips to Go errors.
func vipsError(e C.int) error {
	if e == 0 {
		return nil
	}

	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()
	return errors.New(s)
}

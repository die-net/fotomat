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

var (
	// ErrImageOp means VIPS returned an error but we couldn't get error text.
	ErrImageOp = errors.New("Image operation error")
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

	// The VIPS error buffer is global, and checking and clearing it are
	// not atomic. If errors are infrequent, this will probably return
	// our error. It may also return nothing or unrelated errors.
	// TODO: Consider vips_error_freeze() and skipping this.
	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()

	if s != "" {
		return errors.New(s)
	}

	// At least return something generic.
	return ErrImageOp
}

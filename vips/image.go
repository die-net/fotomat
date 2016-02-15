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

type Image struct {
	image *C.struct__VipsImage
}

func (image Image) Close() {
	C.g_object_unref(C.gpointer(image.image))
	image.image = nil
}

func vipsError(e C.int) error {
	if e == 0 {
		return nil
	}

	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()
	return errors.New(s)
}

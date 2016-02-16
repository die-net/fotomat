package vips

/*
#cgo pkg-config: vips
#include "header.h"
*/
import "C"

const (
	ExifOrientation = "exif-ifd0-Orientation" // Not exposed as a symbol
)

func (in Image) ImageGetAsString(field string) (string, bool) {
	var out *C.char
	e := C.cgo_vips_image_get_as_string(in.vi, C.CString(field), &out)

	s := C.GoString(out)
	C.g_free(C.gpointer(out))

	return s, e == 0
}

func (in Image) ImageRemove(field string) bool {
	ok := C.cgo_vips_image_remove(in.vi, C.CString(field))

	return ok != 0
}

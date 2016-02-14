package vips

/*
#cgo pkg-config: vips
#include "foreign.h"
*/
import "C"

import (
	"unsafe"
)

func JpegloadBuffer(buf []byte, shrink int) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_jpegload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out.image, C.int(shrink)))
	return out, err
}

func PngloadBuffer(buf []byte) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_pngload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out.image))
	return out, err
}

func WebploadBuffer(buf []byte) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_webpload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out.image))
	return out, err
}

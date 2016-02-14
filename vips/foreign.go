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

func (in *VipsImage) JpegsaveBuffer(strip bool, q int, optimizeCoding, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	err := vipsError(C.cgo_vips_jpegsave_buffer(in.image, &ptr, &length, C.int(btoi(strip)), C.int(q), C.int(btoi(optimizeCoding)), C.int(btoi(interlace))))

	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))

	return buf, err
}

func PngloadBuffer(buf []byte) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_pngload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out.image))
	return out, err
}

func (in *VipsImage) PngsaveBuffer(compression int, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	err := vipsError(C.cgo_vips_pngsave_buffer(in.image, &ptr, &length, C.int(compression), C.int(btoi(interlace))))

	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))

	return buf, err
}

func WebploadBuffer(buf []byte) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_webpload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out.image))
	return out, err
}

func (in *VipsImage) WebpsaveBuffer(q int) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	err := vipsError(C.cgo_vips_webpsave_buffer(in.image, &ptr, &length, C.int(q)))

	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))

	return buf, err
}

package vips

/*
#cgo pkg-config: vips
#include "foreign.h"
*/
import "C"

import (
	"unsafe"
)

func Jpegload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload(C.CString(filename), &out, 1)
	return loadError(out, e)
}

func JpegloadShrink(filename string, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload(C.CString(filename), &out, C.int(shrink))
	return loadError(out, e)
}

func JpegloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, 1)
	return loadError(out, e)
}

func JpegloadBufferShrink(buf []byte, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, C.int(shrink))
	return loadError(out, e)
}

func (in *Image) JpegsaveBuffer(strip bool, q int, optimizeCoding, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_jpegsave_buffer(in.vi, &ptr, &length, C.int(btoi(strip)), C.int(q), C.int(btoi(optimizeCoding)), C.int(btoi(interlace)))

	return saveError(ptr, length, e)
}

func Magickload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_magickload(C.CString(filename), &out)
	return loadError(out, e)
}

func MagickloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_magickload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return loadError(out, e)
}

func Pngload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_pngload(C.CString(filename), &out)
	return loadError(out, e)
}

func PngloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_pngload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return loadError(out, e)
}

func (in *Image) PngsaveBuffer(compression int, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_pngsave_buffer(in.vi, &ptr, &length, C.int(compression), C.int(btoi(interlace)))

	return saveError(ptr, length, e)
}

func Webpload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_webpload(C.CString(filename), &out)
	return loadError(out, e)
}

func WebploadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_webpload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return loadError(out, e)
}

func (in *Image) WebpsaveBuffer(q int, lossless bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_webpsave_buffer(in.vi, &ptr, &length, C.int(q), C.int(btoi(lossless)))

	return saveError(ptr, length, e)
}

// loadError is a convenience wrapper around vipsError() for funcs that
// call a vips function passing in an output C.struct__VipsImage and return
// (Image, error).
func loadError(out *C.struct__VipsImage, e C.int) (*Image, error) {
	if err := vipsError(e); err != nil {
		return nil, err
	}

	return imageFromVi(out), nil
}

// saveError is a convenience wrapper around vipsError() for funcs that
// call a vips function passing in an output C buffer and length and returning
// ([]byte, error).
func saveError(ptr unsafe.Pointer, length C.size_t, e C.int) ([]byte, error) {
	err := vipsError(e)

	if ptr == nil {
		return nil, err
	}

	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))

	return buf, err
}

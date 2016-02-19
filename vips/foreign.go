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
	e := C.cgo_vips_jpegload(C.CString(filename), &out)
	return imageError(out, e)
}

func JpegloadShrink(filename string, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload_shrink(C.CString(filename), &out, C.int(shrink))
	return imageError(out, e)
}

func JpegloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return imageError(out, e)
}

func JpegloadBufferShrink(buf []byte, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload_buffer_shrink(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, C.int(shrink))
	return imageError(out, e)
}

func (in *Image) JpegsaveBuffer(strip bool, q int, optimizeCoding, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_jpegsave_buffer(in.vi, &ptr, &length, C.int(btoi(strip)), C.int(q), C.int(btoi(optimizeCoding)), C.int(btoi(interlace)))

	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))

	return buf, vipsError(e)
}

func Magickload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_magickload(C.CString(filename), &out)
	return imageError(out, e)
}

func MagickloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_magickload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return imageError(out, e)
}

func Pngload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_pngload(C.CString(filename), &out)
	return imageError(out, e)
}

func PngloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_pngload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return imageError(out, e)
}

func (in *Image) PngsaveBuffer(compression int, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_pngsave_buffer(in.vi, &ptr, &length, C.int(compression), C.int(btoi(interlace)))

	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))

	return buf, vipsError(e)
}

func Webpload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_webpload(C.CString(filename), &out)
	return imageError(out, e)
}

func WebploadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_webpload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return imageError(out, e)
}

func (in *Image) WebpsaveBuffer(q int, lossless bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_webpsave_buffer(in.vi, &ptr, &length, C.int(q), C.int(btoi(lossless)))

	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))

	return buf, vipsError(e)
}

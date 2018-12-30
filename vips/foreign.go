package vips

/*
#cgo pkg-config: vips
#include "foreign.h"
*/
import "C"

import (
	"unsafe"
)

// Gifload reads a GIF file into an Image.
func Gifload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_gifload(cf, &out)
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// GifloadBuffer reads a GIF byte slice into an Image.
func GifloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_gifload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return loadError(out, e)
}

// Pdfload reads a PDF file into an Image at 72 dpi.
func Pdfload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_pdfload(cf, &out)
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// PdfloadBuffer reads a PDF byte slice into an Image at 72 dpi.
func PdfloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_pdfload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, 1.0)
	return loadError(out, e)
}

// PdfloadBufferShrink reads a PDF byte slice into an Image at (72 /
// shrink) dpi.
func PdfloadBufferShrink(buf []byte, shrink int) (*Image, error) {
	if shrink < 1 {
		shrink = 1
	}
	scale := 1.0 / float64(shrink)

	var out *C.struct__VipsImage
	e := C.cgo_vips_pdfload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, C.double(scale))
	return loadError(out, e)
}

// Jpegload reads and returns a JPEG file as an Image.
func Jpegload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_jpegload(cf, &out, 1)
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// JpegloadShrink reads and returns a JPEG file as an Image, shrinking by an
// integer factor of 1, 2, 4, or 8 during load.  Shrinking during read is
// much faster than decompressing the whole image and then resizing later.
func JpegloadShrink(filename string, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_jpegload(cf, &out, C.int(shrink))
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// JpegloadBuffer reads and returns a JPEG byte slice as an Image.
func JpegloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, 1)
	return loadError(out, e)
}

// JpegloadBufferShrink reads and returns a JPEG byte slice as an Image,
// shrinking by an integer factor of 1, 2, 4, or 8 during load.  Shrinking
// during read is very much faster than decompressing the whole image and
// then shrinking later.
func JpegloadBufferShrink(buf []byte, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_jpegload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, C.int(shrink))
	return loadError(out, e)
}

// JpegsaveBuffer write a VIPS image to a byte slice as JPEG.
// Strip removes all metadata from an image.
// OptimizeCoding computes and uses optimal Huffman coding tables and attaches them.
// Interlace write an interlaced (progressive) JPEG.
func (in *Image) JpegsaveBuffer(strip bool, q int, optimizeCoding, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_jpegsave_buffer(in.vi, &ptr, &length, C.int(btoi(strip)), C.int(q), C.int(btoi(optimizeCoding)), C.int(btoi(interlace)))

	return saveError(ptr, length, e)
}

// Pngload reads a PNG file into an Image.
func Pngload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_pngload(cf, &out)
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// PngloadBuffer reads a PNG byte slice into an Image.
func PngloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_pngload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return loadError(out, e)
}

// PngsaveBuffer write a VIPS image to a byte slice as PNG.
// Strip removes all metadata from an image.
// Compression supplies the gzip level of effort to use (1 - 9).
// Interlace writes the image with ADAM7 interlacing, which is up to 7x slower.
func (in *Image) PngsaveBuffer(strip bool, compression int, interlace bool) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)

	e := C.cgo_vips_pngsave_buffer(in.vi, &ptr, &length, C.int(btoi(strip)), C.int(compression), C.int(btoi(interlace)))

	return saveError(ptr, length, e)
}

// Svgload reads an SVG file into an Image at 72 dpi.
func Svgload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_svgload(cf, &out)
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// SvgloadBuffer reads an SVG byte slice into an Image at 72 dpi.
func SvgloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_svgload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, 1.0)
	return loadError(out, e)
}

// SvgloadBufferShrink reads an SVG byte slice into an Image at (72 /
// shrink) dpi.
func SvgloadBufferShrink(buf []byte, shrink int) (*Image, error) {
	if shrink < 1 {
		shrink = 1
	}
	scale := 1.0 / float64(shrink)

	var out *C.struct__VipsImage
	e := C.cgo_vips_svgload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, C.double(scale))
	return loadError(out, e)
}

// Tiffload reads a TIFF file into an Image.
func Tiffload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_tiffload(cf, &out)
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// TiffloadBuffer reads a TIFF byte slice into an Image.
func TiffloadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_tiffload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out)
	return loadError(out, e)
}

// Webpload read a WebP file into an Image.
func Webpload(filename string) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_webpload(cf, &out, 1)
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// WebploadShrink read a WebP file into an Image, shrinking by an
// integer factor of 1 to 1024 during load.  Shrinking during read is
// much faster than decompressing the whole image and then resizing later.
func WebploadShrink(filename string, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	cf := C.CString(filename)
	e := C.cgo_vips_webpload(cf, &out, C.int(shrink))
	C.free(unsafe.Pointer(cf))
	return loadError(out, e)
}

// WebploadBuffer read a WebP byte slice into an Image.
func WebploadBuffer(buf []byte) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_webpload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, 1)
	return loadError(out, e)
}

// WebploadBufferShrink read a WebP byte slice into an Image, shrinking by an
// integer factor of 1 to 1024 during load.  Shrinking during read is
// much faster than decompressing the whole image and then resizing later.
func WebploadBufferShrink(buf []byte, shrink int) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_webpload_buffer(unsafe.Pointer(&buf[0]), C.size_t(len(buf)), &out, C.int(shrink))
	return loadError(out, e)
}

// WebpsaveBuffer writes an Image to a WebP byte slice.
// Q specifies the compression factor for RGB channels between 0 and 100.
// Lossless encodes the image without any loss, at a large file size.
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

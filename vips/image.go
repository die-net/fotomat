package vips

/*
#cgo pkg-config: vips
#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>
*/
import "C"

type Image struct {
	vi *C.struct__VipsImage
}

func imageFromVi(vi *C.struct__VipsImage) *Image {
	if vi == nil {
		return nil
	}

	return &Image{vi: vi}
}

func (image *Image) Xsize() int {
	return int(image.vi.Xsize)
}

func (image *Image) Ysize() int {
	return int(image.vi.Ysize)
}

func (in *Image) Write() (*Image, error) {
	out := C.vips_image_new_memory()
	e := C.vips_image_write(in.vi, out)
	return imageError(out, e)
}

func (image *Image) Close() {
	C.g_object_unref(C.gpointer(image.vi))
	*image = Image{}
}

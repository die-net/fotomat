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

func (image *Image) Close() {
	C.g_object_unref(C.gpointer(image.vi))
	*image = Image{}
}

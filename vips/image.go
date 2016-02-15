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
	vi     *C.struct__VipsImage
	width  int
	height int
}

func imageFromVi(vi *C.struct__VipsImage) *Image {
	if vi == nil {
		return nil
	}

	image := &Image{
		vi:     vi,
		width:  int(vi.Xsize),
		height: int(vi.Ysize),
	}
	return image
}

func (image *Image) Close() {
	C.g_object_unref(C.gpointer(image.vi))
	*image = Image{}
}

func vipsError(e C.int) error {
	if e == 0 {
		return nil
	}

	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()
	return errors.New(s)
}

func imageError(out *C.struct__VipsImage, e C.int) (*Image, error) {
	if err := vipsError(e); err != nil {
		return nil, err
	}

	return imageFromVi(out), nil
}

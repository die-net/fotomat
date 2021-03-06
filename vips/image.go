package vips

/*
#cgo pkg-config: vips
#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>
*/
import "C"

// Image can represent an image on disc, a memory buffer, or a partially
// evaluated image in memory, represented as its source data and chain of
// operations to be performed on that image later.
type Image struct {
	vi *C.struct__VipsImage
}

func imageFromVi(vi *C.struct__VipsImage) *Image {
	if vi == nil {
		return nil
	}

	return &Image{vi: vi}
}

// Xsize returns the width of the image in pixels.
func (in *Image) Xsize() int {
	return int(in.vi.Xsize)
}

// Ysize returns the height of the image in pixels.
func (in *Image) Ysize() int {
	return int(in.vi.Ysize)
}

// Write applies all queued operations to the source image copies the result
// to a new memory buffer.
func (in *Image) Write() error {
	out := C.vips_image_new_memory()
	e := C.vips_image_write(in.vi, out)
	return in.imageError(out, e)
}

// Close frees the memory associated with an Image.
func (in *Image) Close() {
	C.g_object_unref(C.gpointer(in.vi))
	*in = Image{}
}

// imageError adapts image modification semantics from being the VIPS-style
// chain of immutable objects that we need to individually free to a single
// stateful Go object that can be closed once, greatly simplifying error
// handling.
func (in *Image) imageError(out *C.struct__VipsImage, e C.int) error {
	if in.vi == nil {
		panic("Input image is nil")
	}

	if err := vipsError(e); err != nil {
		if out != nil {
			C.g_object_unref(C.gpointer(out))
		}
		return err
	}

	C.g_object_unref(C.gpointer(in.vi))
	in.vi = out

	return nil
}

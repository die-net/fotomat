// Copyright 2013 Herbert G. Fischer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package imagick

/*
#cgo pkg-config: MagickWand MagickCore
#include <wand/MagickWand.h>
*/
import "C"

import (
	"unsafe"
)

// This struct represents the MagickWand C API of ImageMagick
type MagickWand struct {
	mw *C.MagickWand
}

// Returns a wand required for all other methods in the API. A fatal exception is thrown if there is not enough memory to allocate the wand.
// Use Destroy() to dispose of the wand then it is no longer needed.
func NewMagickWand() *MagickWand {
	return &MagickWand{C.NewMagickWand()}
}

// Returns a wand with an image/
func NewMagickWandFromImage(img *Image) *MagickWand {
	return &MagickWand{C.NewMagickWandFromImage(img.img)}
}

// Clear resources associated with the wand, leaving the wand blank, and ready to be used for a new set of images.
func (mw *MagickWand) Clear() {
	C.ClearMagickWand(mw.mw)
}

// Makes an exact copy of the MagickWand object
func (mw *MagickWand) Clone() *MagickWand {
	clone := C.CloneMagickWand(mw.mw)
	return &MagickWand{clone}
}

// Deallocates memory associated with an MagickWand
func (mw *MagickWand) Destroy() {
	if mw.mw == nil {
		return
	}
	mw.mw = C.DestroyMagickWand(mw.mw)
	C.free(unsafe.Pointer(mw.mw))
	mw.mw = nil
}

// Returns true if the wand is a verified magick wand
func (mw *MagickWand) IsVerified() bool {
	if mw.mw != nil {
		return 1 == C.int(C.IsMagickWand(mw.mw))
	}
	return false
}

// Relinquishes memory resources returned by such methods as MagickIdentifyImage(), MagickGetException(), etc
func (mw *MagickWand) relinquishMemory(ptr unsafe.Pointer) {
	C.MagickRelinquishMemory(ptr)
}

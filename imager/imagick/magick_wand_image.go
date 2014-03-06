// Copyright 2013 Herbert G. Fischer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package imagick

/*
#include <wand/MagickWand.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Returns the current image from the magick wand
func (mw *MagickWand) GetImageFromMagickWand() *Image {
	return &Image{C.GetImageFromMagickWand(mw.mw)}
}

// Extracts a region of the image
func (mw *MagickWand) CropImage(width, height uint, x, y int) error {
	C.MagickCropImage(mw.mw, C.ulong(width), C.ulong(height), C.long(x), C.long(y))
	return mw.GetLastError()
}

// Dereferences an image, deallocating memory associated with the image if the
// reference count becomes zero.
func (mw *MagickWand) DestroyImage(img *Image) *Image {
	return &Image{C.MagickDestroyImage(img.img)}
}

// Implements direct to memory image formats. It returns the image as a blob
// (a formatted "file" in memory) and its length, starting from the current
// position in the image sequence. Use SetImageFormat() to set the format to
// write to the blob (GIF, JPEG, PNG, etc.). Utilize ResetIterator() to ensure
// the write is from the beginning of the image sequence.
func (mw *MagickWand) GetImageBlob() []byte {
	clen := C.size_t(0)
	csblob := C.MagickGetImageBlob(mw.mw, &clen)
	defer mw.relinquishMemory(unsafe.Pointer(csblob))
	return C.GoBytes(unsafe.Pointer(csblob), C.int(clen))
}

// Returns the format of a particular image in a sequence.
func (mw *MagickWand) GetImageFormat() string {
	return C.GoString(C.MagickGetImageFormat(mw.mw))
}

// Returns the image height.
func (mw *MagickWand) GetImageHeight() uint {
	return uint(C.MagickGetImageHeight(mw.mw))
}

// Returns the image width.
func (mw *MagickWand) GetImageWidth() uint {
	return uint(C.MagickGetImageWidth(mw.mw))
}

// Pings an image or image sequence from a blob.
func (mw *MagickWand) PingImageBlob(blob []byte) error {
	C.MagickPingImageBlob(mw.mw, unsafe.Pointer(&blob[0]), C.size_t(len(blob)))
	return mw.GetLastError()
}

// Reads an image or image sequence from a blob.
func (mw *MagickWand) ReadImageBlob(blob []byte) error {
	if len(blob) == 0 {
		return errors.New("zero-length blob not permitted")
	}
	C.MagickReadImageBlob(mw.mw, unsafe.Pointer(&blob[0]), C.size_t(len(blob)))
	return mw.GetLastError()
}


// Scales an image to the desired dimensions
//
// cols: the number of cols in the scaled image.
//
// rows: the number of rows in the scaled image.
//
// filter: Image filter to use.
//
// blur: the blur factor where > 1 is blurry, < 1 is sharp.
//
func (mw *MagickWand) ResizeImage(cols, rows uint, filter FilterType, blur float64) error {
	C.MagickResizeImage(mw.mw, C.ulong(cols), C.ulong(rows), C.FilterTypes(filter), C.double(blur))
	return mw.GetLastError()
}

// Activates, deactivates, resets, or sets the alpha channel.
func (mw *MagickWand) SetImageAlphaChannel(act AlphaChannelType) error {
	C.MagickSetImageAlphaChannel(mw.mw, C.AlphaChannelType(act))
	return mw.GetLastError()
}

// Sets the image compression.
func (mw *MagickWand) SetImageCompression(compression CompressionType) error {
	C.MagickSetImageCompression(mw.mw, C.CompressionType(compression))
	return mw.GetLastError()
}

// Sets the image compression quality.
func (mw *MagickWand) SetImageCompressionQuality(quality uint) error {
	C.MagickSetImageCompressionQuality(mw.mw, C.ulong(quality))
	return mw.GetLastError()
}

// Sets the image depth.
//
// depth: the image depth in bits: 8, 16, or 32.
//
func (mw *MagickWand) SetImageDepth(depth uint) error {
	C.MagickSetImageDepth(mw.mw, C.ulong(depth))
	return mw.GetLastError()
}

// Sets the format of a particular image in a sequence.
//
// format: the image format.
//
func (mw *MagickWand) SetImageFormat(format string) error {
	csformat := C.CString(format)
	defer C.free(unsafe.Pointer(csformat))
	C.MagickSetImageFormat(mw.mw, csformat)
	return mw.GetLastError()
}


// Sets the image interlace scheme.
func (mw *MagickWand) SetImageInterlaceScheme(interlace InterlaceType) error {
	C.MagickSetImageInterlaceScheme(mw.mw, C.InterlaceType(interlace))
	return mw.GetLastError()
}

// Unsharpens an image. We convolve the image with a Gaussian operator of the
// given radius and standard deviation (sigma). For reasonable results, radius
// should be larger than sigma. Use a radius of 0 and UnsharpMaskImage()
// selects a suitable radius for you.
//
// radius: the radius of the Gaussian, in pixels, not counting the center pixel.
//
// sigma: the standard deviation of the Gaussian, in pixels.
//
// amount: the percentage of the difference between the original and the blur
// image that is added back into the original.
//
// threshold: the threshold in pixels needed to apply the diffence amount.
//
func (mw *MagickWand) UnsharpMaskImage(radius, sigma, amount, threshold float64) error {
	C.MagickUnsharpMaskImage(mw.mw, C.double(radius), C.double(sigma), C.double(amount), C.double(threshold))
	return mw.GetLastError()
}

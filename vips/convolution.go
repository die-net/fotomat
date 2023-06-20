package vips

/*
#cgo pkg-config: vips
#include "convolution.h"
*/
import "C"

// Gaussblur creates a circularly symmetric Gaussian mask of radius sigma
// and performs a separable (two-pass) convolution of in with it.
func (in *Image) Gaussblur(sigma float64) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_gaussblur(in.vi, &out, C.double(sigma))
	return in.imageError(out, e)
}

// PhotoMetric takes a histogram of a Sobel edge detect of our image.
// Returns the highest number of histogram values in a row that are more
// than the maximum value * threshold.  With a threshold of 0.01, more than
// 16 indicates a photo.
func (in *Image) PhotoMetric(threshold float64) (int, error) {
	var out C.int
	err := vipsError(C.cgo_photo_metric(in.vi, C.double(threshold), &out))
	return int(out), err
}

// MildSharpen performs a fast, mild sharpen of an Image.
func (in *Image) MildSharpen() error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_mild_sharpen(in.vi, &out)
	return in.imageError(out, e)
}

// Sharpen performs a gaussian blur of radius and subtracts from in to
// generate a high-frequency signal.  This signal is passed through a lookup
// table generated from the parameters (x1: flat/jaggy threshold, y2:
// maximum amount of brightening, y3: maximum amount of darkening, m1: slope
// for flat areas, m2: slope for jaggy areas) and added back to in.
func (in *Image) Sharpen(radius int, x1, y2, y3, m1, m2 float64) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_sharpen(in.vi, &out, C.int(radius), C.double(x1), C.double(y2), C.double(y3), C.double(m1), C.double(m2))
	return in.imageError(out, e)
}

// Sobel is an edge detection filter that converts in to black and white and
// subtracts pixels from their neighbors.  The values that remain are higher
// when the brightness of a given pixel differs greatly from its neighbors.
func (in *Image) Sobel() error {
	var out *C.struct__VipsImage
	e := C.cgo_sobel(in.vi, &out)
	return in.imageError(out, e)
}

package vips

/*
#cgo pkg-config: vips
#include "colour.h"
*/
import "C"

import (
	"unsafe"
)

// Interpretation suggests how the values in an image should be interpreted.
// For example, a three-band float image of type InterpretationLAB
// should have its pixels interpreted as coordinates in CIE Lab space.
type Interpretation int

// Various Interpretation values understood by VIPS.
const (
	InterpretationMultiband Interpretation = C.VIPS_INTERPRETATION_MULTIBAND
	InterpretationBW        Interpretation = C.VIPS_INTERPRETATION_B_W
	InterpretationHistogram Interpretation = C.VIPS_INTERPRETATION_HISTOGRAM
	InterpretationXYZ       Interpretation = C.VIPS_INTERPRETATION_XYZ
	InterpretationLAB       Interpretation = C.VIPS_INTERPRETATION_LAB
	InterpretationCMYK      Interpretation = C.VIPS_INTERPRETATION_CMYK
	InterpretationLABQ      Interpretation = C.VIPS_INTERPRETATION_LABQ
	InterpretationRGB       Interpretation = C.VIPS_INTERPRETATION_RGB
	InterpretationCMC       Interpretation = C.VIPS_INTERPRETATION_CMC
	InterpretationLCH       Interpretation = C.VIPS_INTERPRETATION_LCH
	InterpretationLABS      Interpretation = C.VIPS_INTERPRETATION_LABS
	InterpretationSRGB      Interpretation = C.VIPS_INTERPRETATION_sRGB
	InterpretationYXY       Interpretation = C.VIPS_INTERPRETATION_YXY
	InterpretationFourier   Interpretation = C.VIPS_INTERPRETATION_FOURIER
	InterpretationRGB16     Interpretation = C.VIPS_INTERPRETATION_RGB16
	InterpretationGrey16    Interpretation = C.VIPS_INTERPRETATION_GREY16
	InterpretationMatrix    Interpretation = C.VIPS_INTERPRETATION_MATRIX
	InterpretationScRGB     Interpretation = C.VIPS_INTERPRETATION_scRGB
)

// Intent is the color management system rendering intent.
type Intent int

// Various Intent values understood by VIPS.
const (
	IntentPerceptual Intent = C.VIPS_INTENT_PERCEPTUAL // best for business graphics
	IntentRelative   Intent = C.VIPS_INTENT_RELATIVE   // best for accurate communication with other imaging libraries
	IntentSaturation Intent = C.VIPS_INTENT_SATURATION // best for business graphics
	IntentAbsolute   Intent = C.VIPS_INTENT_ABSOLUTE   // best for scientific work
)

// Colourspace moves an image to a target colourspace using the best sequence of colour transform operations.
func (in *Image) Colourspace(space Interpretation) error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_colourspace(in.vi, &out, C.VipsInterpretation(space))
	return in.imageError(out, e)
}

// IccImport moves an image from device space to D65 LAB using the image's
// embedded ICC profile.
func (in *Image) IccImport() error {
	var out *C.struct__VipsImage
	e := C.cgo_vips_icc_import(in.vi, &out)
	return in.imageError(out, e)
}

// IccTransform transform an image with a pair of ICC profiles. The input
// image is moved to profile-connection space with the input profile and
// then to the output space with the output profile and intent.
func (in *Image) IccTransform(outputProfileFilename, inputProfileFilename string, intent Intent) error {
	var out *C.struct__VipsImage
	co := C.CString(outputProfileFilename)
	var ci *C.char
	if inputProfileFilename != "" {
		ci = C.CString(inputProfileFilename)
	}
	e := C.cgo_vips_icc_transform(in.vi, &out, co, ci, C.VipsIntent(intent))
	C.free(unsafe.Pointer(ci))
	C.free(unsafe.Pointer(co))
	return in.imageError(out, e)
}

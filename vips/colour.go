package vips

/*
#cgo pkg-config: vips
#include "colour.h"
*/
import "C"

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

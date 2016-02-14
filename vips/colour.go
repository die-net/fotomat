package vips

/*
#cgo pkg-config: vips
#include "colour.h"
*/
import "C"

type Interpretation int

const (
	InterpretationMultiband = C.VIPS_INTERPRETATION_MULTIBAND
	InterpretationBW        = C.VIPS_INTERPRETATION_B_W
	InterpretationHistogram = C.VIPS_INTERPRETATION_HISTOGRAM
	InterpretationXYZ       = C.VIPS_INTERPRETATION_XYZ
	InterpretationLAB       = C.VIPS_INTERPRETATION_LAB
	InterpretationCMYK      = C.VIPS_INTERPRETATION_CMYK
	InterpretationLABQ      = C.VIPS_INTERPRETATION_LABQ
	InterpretationRGB       = C.VIPS_INTERPRETATION_RGB
	InterpretationCMC       = C.VIPS_INTERPRETATION_CMC
	InterpretationLCH       = C.VIPS_INTERPRETATION_LCH
	InterpretationLABS      = C.VIPS_INTERPRETATION_LABS
	InterpretationSRGB      = C.VIPS_INTERPRETATION_sRGB
	InterpretationYXY       = C.VIPS_INTERPRETATION_YXY
	InterpretationFourier   = C.VIPS_INTERPRETATION_FOURIER
	InterpretationRGB16     = C.VIPS_INTERPRETATION_RGB16
	InterpretationGrey16    = C.VIPS_INTERPRETATION_GREY16
	InterpretationMatrix    = C.VIPS_INTERPRETATION_MATRIX
	InterpretationScRGB     = C.VIPS_INTERPRETATION_scRGB
	InterpretationHSV       = C.VIPS_INTERPRETATION_HSV
)

func (in VipsImage) Colourspace(space Interpretation) (VipsImage, error) {
	out := VipsImage{}
	err := vipsError(C.cgo_vips_colourspace(in.image, &out.image, C.VipsInterpretation(space)))
	return out, err
}

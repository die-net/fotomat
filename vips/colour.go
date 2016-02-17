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
)

// Colourspace moves an image to a target colourspace using the best sequence of colour transform operations.
func (in Image) Colourspace(space Interpretation) (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_colourspace(in.vi, &out, C.VipsInterpretation(space))
	return imageError(out, e)
}

func (in Image) IccImport() (*Image, error) {
	var out *C.struct__VipsImage
	e := C.cgo_vips_icc_import(in.vi, &out)
	return imageError(out, e)
}

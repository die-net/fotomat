package vips

/*
#cgo pkg-config: vips
#include "header.h"
*/
import "C"

const (
	ExifOrientation = "exif-ifd0-Orientation" // Not exposed as a symbol
	MetaIccName     = "icc-profile-data"
)

const (
	FormatNotSet    = C.VIPS_FORMAT_NOTSET
	FormatUchar     = C.VIPS_FORMAT_UCHAR
	FormatChar      = C.VIPS_FORMAT_CHAR
	FormatUshort    = C.VIPS_FORMAT_USHORT
	FormatShort     = C.VIPS_FORMAT_SHORT
	FormatUint      = C.VIPS_FORMAT_UINT
	FormatInt       = C.VIPS_FORMAT_INT
	FormatFloaT     = C.VIPS_FORMAT_FLOAT
	FormatComplex   = C.VIPS_FORMAT_COMPLEX
	FormatDouble    = C.VIPS_FORMAT_DOUBLE
	FormatDpComplex = C.VIPS_FORMAT_DPCOMPLEX
)

func (in Image) ImageGetAsString(field string) (string, bool) {
	var out *C.char
	e := C.cgo_vips_image_get_as_string(in.vi, C.CString(field), &out)

	s := C.GoString(out)
	// TODO: Leak? Crash if I follow docs and: C.g_free(C.gpointer(out))

	return s, e == 0
}

func (in Image) ImageGetBands() int {
	return int(C.vips_image_get_bands(in.vi))
}

func (in Image) ImageGetFormat() int {
	return int(C.vips_image_get_format(in.vi))
}

func (in Image) ImageGuessInterpretation() int {
	return int(C.vips_image_guess_interpretation(in.vi))
}

func (in Image) HasAlpha() bool {
	b := in.ImageGetBands()
	i := in.ImageGuessInterpretation()

	alpha := (b == 2 && i == InterpretationBW) ||
		(b == 4 && i != InterpretationCMYK) ||
		(b == 5 && i == InterpretationCMYK)
	return alpha
}

func (in Image) ImageRemove(field string) bool {
	ok := C.vips_image_remove(in.vi, C.CString(field))

	return ok != 0
}

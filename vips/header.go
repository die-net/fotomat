package vips

/*
#cgo pkg-config: vips
#include "header.h"
*/
import "C"

import (
	"unsafe"
)

// Potential values for ImageGetAsString.
const (
	ExifOrientation = "exif-ifd0-Orientation"
	MetaIccName     = "icc-profile-data"
)

// BandFormat is the format used for each band element.  Each corresponds to
// a native C type for the current machine.
type BandFormat int

// Potential values that GetBandFormat will return.
const (
	BandFormatNotSet    BandFormat = C.VIPS_FORMAT_NOTSET
	BandFormatUchar     BandFormat = C.VIPS_FORMAT_UCHAR
	BandFormatChar      BandFormat = C.VIPS_FORMAT_CHAR
	BandFormatUshort    BandFormat = C.VIPS_FORMAT_USHORT
	BandFormatShort     BandFormat = C.VIPS_FORMAT_SHORT
	BandFormatUint      BandFormat = C.VIPS_FORMAT_UINT
	BandFormatInt       BandFormat = C.VIPS_FORMAT_INT
	BandFormatFloat     BandFormat = C.VIPS_FORMAT_FLOAT
	BandFormatComplex   BandFormat = C.VIPS_FORMAT_COMPLEX
	BandFormatDouble    BandFormat = C.VIPS_FORMAT_DOUBLE
	BandFormatDpComplex BandFormat = C.VIPS_FORMAT_DPCOMPLEX
)

func (in *Image) ImageFieldExists(field string) bool {
	cf := C.CString(field)
	e := C.vips_image_get_typeof(in.vi, cf)
	C.free(unsafe.Pointer(cf))

	return e != 0
}

func (in *Image) ImageGetAsString(field string) (string, bool) {
	var out *C.char
	cf := C.CString(field)
	e := C.cgo_vips_image_get_as_string(in.vi, cf, &out)
	C.free(unsafe.Pointer(cf))

	s := C.GoString(out)
	// TODO: Leak? Crash if I follow docs and: C.g_free(C.gpointer(out))

	return s, e == 0
}

func (in *Image) ImageGetBands() int {
	return int(C.vips_image_get_bands(in.vi))
}

func (in *Image) ImageGetBandFormat() BandFormat {
	return BandFormat(C.vips_image_get_format(in.vi))
}

func (in *Image) ImageGuessInterpretation() Interpretation {
	return Interpretation(C.vips_image_guess_interpretation(in.vi))
}

func (in *Image) HasAlpha() bool {
	b := in.ImageGetBands()
	i := in.ImageGuessInterpretation()

	alpha := (b == 2 && i == InterpretationBW) ||
		(b == 4 && i != InterpretationCMYK) ||
		(b == 5 && i == InterpretationCMYK)
	return alpha
}

func (in *Image) ImageRemove(field string) bool {
	cf := C.CString(field)
	ok := C.vips_image_remove(in.vi, cf)
	C.free(unsafe.Pointer(cf))

	return ok != 0
}

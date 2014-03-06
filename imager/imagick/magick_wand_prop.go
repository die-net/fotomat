// Copyright 2013 Herbert G. Fischer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package imagick

/*
#include <wand/MagickWand.h>
*/
import "C"
import (
	//"fmt"
	"unsafe"
)

// This method deletes a image property
func (mw *MagickWand) DeleteImageProperty(property string) error {
	csproperty := C.CString(property)
	defer C.free(unsafe.Pointer(csproperty))
	C.MagickDeleteImageProperty(mw.mw, csproperty)
	return mw.GetLastError()
}

// This method deletes a wand option
func (mw *MagickWand) DeleteOption(option string) error {
	csoption := C.CString(option)
	defer C.free(unsafe.Pointer(csoption))
	C.MagickDeleteOption(mw.mw, csoption)
	return mw.GetLastError()
}

// Returns all the profile names that match the specified pattern associated
// with a wand. Use GetImageProfile() to return the value of a particular
// property.
func (mw *MagickWand) GetImageProfiles(pattern string) (profiles []string) {
	cspattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cspattern))
	np := C.ulong(0)
	ps := C.MagickGetImageProfiles(mw.mw, cspattern, &np)
	profiles = sizedCStringArrayToStringSlice(ps, np)
	return
}

// Gets the wand interlace scheme.
func (mw *MagickWand) GetInterlaceScheme() InterlaceType {
	return InterlaceType(C.MagickGetInterlaceScheme(mw.mw))
}

// Returns a value associated with a wand and the specified key.
func (mw *MagickWand) GetOption(key string) string {
	cskey := C.CString(key)
	defer C.free(unsafe.Pointer(cskey))
	csval := C.MagickGetOption(mw.mw, cskey)
	return C.GoString(csval)
}

// Returns all the option names that match the specified pattern associated
// with a wand. Use GetOption() to return the value of a particular option.
func (mw *MagickWand) GetOptions(pattern string) (options []string) {
	cspattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cspattern))
	np := C.ulong(0)
	ps := C.MagickGetOptions(mw.mw, cspattern, &np)
	options = sizedCStringArrayToStringSlice(ps, np)
	return
}

// Removes the named image profile and returns it.
//
// name: name of profile to return: ICC, IPTC, or generic profile.
//
func (mw *MagickWand) RemoveImageProfile(name string) []byte {
	csname := C.CString(name)
	defer C.free(unsafe.Pointer(csname))
	clen := C.size_t(0)
	profile := C.MagickRemoveImageProfile(mw.mw, csname, &clen)
	return C.GoBytes(unsafe.Pointer(profile), C.int(clen))
}

// Sets the image interlacing scheme
func (mw *MagickWand) SetInterlaceScheme(scheme InterlaceType) error {
	C.MagickSetInterlaceScheme(mw.mw, C.InterlaceType(scheme))
	return mw.GetLastError()
}

// Associates one or options with the wand (.e.g
// SetOption(wand, "jpeg:perserve", "yes")).
func (mw *MagickWand) SetOption(key, value string) error {
	cskey := C.CString(key)
	defer C.free(unsafe.Pointer(cskey))
	csvalue := C.CString(value)
	defer C.free(unsafe.Pointer(csvalue))
	C.MagickSetOption(mw.mw, cskey, csvalue)
	return mw.GetLastError()
}

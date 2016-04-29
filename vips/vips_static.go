// Add appropriate flags for linking VIPS statically into Fotomat if the
// "-tags vips_static" build flag is used.  Configure VIPS with:
// CFLAGS="-fPIC" CXXFLAGS="-fPIC" LDFLAGS="-lstdc++"

// +build vips_static

package vips

/*
#cgo pkg-config: --static vips
#cgo LDFLAGS: -lstdc++ -lgif -ltiff
*/
import "C"

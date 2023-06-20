// Add appropriate flags for linking VIPS statically into Fotomat if the
// "-tags vips_static" build flag is used.  Configure VIPS with:
// CFLAGS="-fPIE" CXXFLAGS="-fPIE" LDFLAGS="-lstdc++"

//go:build vips_static
// +build vips_static

package vips

/*
#cgo pkg-config: --static vips
#cgo LDFLAGS: -lstdc++ -lgif -ltiff -lexpat
*/
import "C"

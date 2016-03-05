// Add appropriate linker flags for VIPS dependencies for making a static
// binary if the "static" build tag is used, such as: go run -tags static

// +build static

package vips

/*
#cgo pkg-config: --static vips
#cgo LDFLAGS: -static
*/
import "C"

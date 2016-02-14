package vips

/*
#cgo pkg-config: vips
#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>
*/
import "C"

import (
	"runtime"
)

func init() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_init(C.CString("govips")); err != 0 {
		C.vips_shutdown()
		panic("vips_initialize error")
	}

	C.vips_concurrency_set(1)
	C.vips_cache_set_max_mem(100 * 1024 * 1024)
	C.vips_cache_set_max(500)
}

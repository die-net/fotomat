package vips

/*
#cgo pkg-config: vips
#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>
#include "init.h"
*/
import "C"

import (
	"runtime"
)

func init() {
	Initialize()
}

func Initialize() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.cgo_vips_init(); err != 0 {
		C.vips_shutdown()
		panic("vips_initialize error")
	}

	C.vips_concurrency_set(1)
	C.vips_cache_set_max_mem(100 * 1024 * 1024)
	C.vips_cache_set_max(500)
}

func ThreadShutdown() {
	C.vips_thread_shutdown()
}

func Shutdown() {
	C.vips_shutdown()
}

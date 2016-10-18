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
	"os"
	"runtime"
)

var (
	ResizeOffset = 0.0
)

// Initialize starts up the world of VIPS. You should call this on program
// startup before using any other VIPS operations.
func Initialize() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	_ = os.Setenv("VIPS_WARNING", "disable")

	if err := C.cgo_vips_init(); err != 0 {
		C.vips_shutdown()
		panic("vips_initialize error")
	}

	C.vips_concurrency_set(1)
	C.vips_cache_set_max_mem(0)
	C.vips_cache_set_max(0)

	// Vips 8.3 sometimes produces 1px smaller images than desired without rounding help.
	if C.VIPS_MAJOR_VERSION == 8 && C.VIPS_MINOR_VERSION < 4 {
		ResizeOffset = 0.5
	}
}

func LeakSet(enable bool) {
	C.vips_leak_set(C.gboolean(btoi(enable)))
}

// ThreadShutdown frees any thread-private data and flushes any profiling
// information.  This function needs to be called when a thread that has
// been using vips exits or there will be memory leaks.  It may be called
// many times, and you can continue using vips after calling it.  Calling it
// too often will reduce performance.
func ThreadShutdown() {
	C.vips_thread_shutdown()
}

// Shutdown drops caches and closes plugins, and runs a leak check if
// requested.  May be called many times.
func Shutdown() {
	C.vips_shutdown()
}

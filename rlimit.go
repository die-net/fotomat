// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"syscall"
)

var (
	maxConnections = flag.Int("max_connections", getRlimitMax(syscall.RLIMIT_NOFILE), "The maximum number of incoming connections allowed.")
)

func getRlimitMax(resource int) int {
	var rlimit syscall.Rlimit

	err := syscall.Getrlimit(resource, &rlimit)

	if err == nil {
		return int(rlimit.Max)
	} else {
		return 0
	}
}

func setRlimit(resource int, value int) {
	rlimit := &syscall.Rlimit{Cur: uint64(value), Max: uint64(value)}

	err := syscall.Setrlimit(resource, rlimit)
	if err != nil {
		log.Fatalln("Error Setting Rlimit ", err)
	}
}

func setRlimitFromFlags() {
	setRlimit(syscall.RLIMIT_NOFILE, *maxConnections)
}

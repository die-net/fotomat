package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" // Adds http://*/debug/pprof/ to default mux.
	"runtime"
)

var (
	listenAddr = flag.String("listen", "127.0.0.1:3520", "[IP]:port to listen for incoming connections.")

	postFns []func()
)

func main() {
	flag.Parse()

	// Allow more threads than that for networking, etc.
	runtime.GOMAXPROCS(*maxImageThreads * 2)

	setupTempdir()

	postRun()

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func post(fn func()) {
	postFns = append(postFns, fn)
}

// Run everything queued with post().
func postRun() {
	for _, fn := range postFns {
		fn()
	}
}

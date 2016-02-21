package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" // Adds http://*/debug/pprof/ to default mux.
	"runtime"
)

var (
	listenAddr      = flag.String("listen", "127.0.0.1:3520", "[IP]:port to listen for incoming connections.")
	maxImageThreads = flag.Int("max_image_threads", runtime.NumCPU(), "Maximum number of threads simultaneously processing images.")
)

func main() {
	flag.Parse()

	// Up to max_threads will be allowed to be blocked in ImageMagick.
	poolInit(*maxImageThreads)

	// Allow more threads than that for networking, etc.
	runtime.GOMAXPROCS(*maxImageThreads * 2)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

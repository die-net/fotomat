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
	maxThreads = flag.Int("max_threads", runtime.NumCPU(), "Maximum number of OS threads to create.")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(*maxThreads)

	log.Printf("Listening on %s\n", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

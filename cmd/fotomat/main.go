package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // Adds http://*/debug/pprof/ to default mux.
	"os"
	"runtime"
)

var (
	listenAddr = flag.String("listen", "127.0.0.1:3520", "[IP]:port to listen for incoming connections.")
	version    = flag.Bool("version", false, "Show version and exit.")
)

func main() {
	flag.Parse()

	// Allow more threads than that for networking, etc.
	runtime.GOMAXPROCS(*maxImageThreads * 2)

	rlimitInit()

	handleInit()

	if *version {
		fmt.Println("Fotomat v" + FotomatVersion)
		os.Exit(0)
	}

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

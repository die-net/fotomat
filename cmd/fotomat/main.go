package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: Move this to its own port.
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	debugListen = flag.String("debug_listen", "127.0.0.1:3521", "[IP]:port to listen for pprof and metrics requests. (\"\" = disable)")
	listen      = flag.String("listen", "127.0.0.1:3520", "[IP]:port to listen for image serving requests.")
	version     = flag.Bool("version", false, "Show version and exit.")
)

func main() {
	flag.Parse()

	// Allow more threads than that for networking, etc.
	runtime.GOMAXPROCS(*maxImageThreads * 2)

	rlimitInit()
	prometheusInit()

	if *version {
		fmt.Println("Fotomat v" + FotomatVersion)
		os.Exit(0)
	}

	if *debugListen != "" {
		go func() {
			ps := &http.Server{
				Addr:         *debugListen,
				Handler:      promhttp.Handler(),
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  5 * time.Minute,
			}
			log.Fatal(ps.ListenAndServe())
		}()
	}

	handler := handleInit()
	srv := &http.Server{
		Addr:         *listen,
		Handler:      prometheusWrapHandler(handler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}
	log.Fatal(srv.ListenAndServe())
}

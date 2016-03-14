package main

import (
	"flag"
	"fmt"
	"github.com/die-net/fotomat/format"
	"github.com/die-net/fotomat/thumbnail"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	fetchTimeout        = flag.Duration("fetch_timeout", 30*time.Second, "How long to wait to receive original image from source (0=disable).")
	localImageDirectory = flag.String("local_image_directory", "", "Enable local image serving from this path (\"\"=proxy instead).")
	maxImageThreads     = flag.Int("max_image_threads", numCpuCores(), "Maximum number of threads simultaneously processing images (0=all CPUs).")
	maxPrefetch         = flag.Int("max_prefetch", numCpuCores(), "Maximum number of images to prefetch before thread is available.")

	transport      http.Transport
	client         http.Client
	pool           *thumbnail.Pool
	fetchSemaphore chan bool
)

func handlerInit() {
	http.HandleFunc("/", imageProxyHandler)

	transport = http.Transport{Proxy: http.ProxyFromEnvironment}
	if *localImageDirectory != "" {
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(*localImageDirectory)))
	}

	client = http.Client{Transport: http.RoundTripper(&transport), Timeout: *fetchTimeout}

	pool = thumbnail.NewPool(*maxImageThreads, 1)

	limit := *maxImageThreads + *maxPrefetch
	fetchSemaphore = make(chan bool, limit)
	for i := 0; i < limit; i++ {
		fetchSemaphore <- true
	}
}

func imageProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		sendError(w, nil, http.StatusMethodNotAllowed)
		return
	}

	path, options, saveOptions, ok := pathParse(r.URL.Path)
	if !ok {
		sendError(w, nil, 400)
		return
	}

	var u *url.URL
	if *localImageDirectory == "" {
		u = &url.URL{Scheme: "http", Host: r.Host, Path: path}
	} else {
		u = &url.URL{Scheme: "file", Host: "localhost", Path: path}
	}

	fetchAndProcessImage(w, u.String(), options, saveOptions)
}

var userAgent = "Fotomat/" + FotomatVersion + " (https://github.com/die-net/fotomat)"

func fetchAndProcessImage(w http.ResponseWriter, url string, options thumbnail.Options, saveOptions format.SaveOptions) {
	aborted := w.(http.CloseNotifier).CloseNotify()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendError(w, err, 0)
	}

	req.Header.Set("User-Agent", userAgent)

	// Wait for our turn to fetch and hold the original image.
	<-fetchSemaphore

	resp, err := client.Do(req)
	if err != nil {
		fetchSemaphore <- true // Free up ASAP.
		sendError(w, err, 0)
		return
	}

	// Has client closed connection while we were waiting?
	select {
	case <-aborted:
		fetchSemaphore <- true // Free up ASAP.
		sendError(w, nil, http.StatusRequestTimeout)
		return
	default:
	}

	orig, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		fetchSemaphore <- true // Free up ASAP.
		resp.Body.Close()
		sendError(w, err, resp.StatusCode)
		return
	}

	resp.Body.Close()

	thumb, err := pool.Thumbnail(orig, options, saveOptions, aborted)
	orig = nil // Free up image memory ASAP.

	fetchSemaphore <- true // Free up ASAP.

	if err != nil {
		sendError(w, err, 0)
		return
	}

	w.Header().Set("Server", "Fotomat")
	w.Header().Set("Content-Length", strconv.Itoa(len(thumb)))
	w.Write(thumb)
}

func sendError(w http.ResponseWriter, err error, status int) {
	switch status {
	case http.StatusNoContent,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusRequestTimeout,
		http.StatusGone:
		err = nil
	case 0:
		switch err {
		case format.ErrUnknownFormat, thumbnail.ErrTooSmall:
			status = http.StatusUnsupportedMediaType
		case thumbnail.ErrTooBig:
			status = http.StatusRequestEntityTooLarge
		default:
			if isTimeout(err) {
				err = nil
				status = http.StatusGatewayTimeout
			} else {
				status = http.StatusInternalServerError
			}
		}
	default:
		err = fmt.Errorf("Proxy received %d %s", status, http.StatusText(status))
		status = http.StatusBadGateway
	}
	if err == nil {
		err = fmt.Errorf(http.StatusText(status))
	}
	http.Error(w, err.Error(), status)
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case net.Error:
		return err.Timeout()
	case *url.Error:
		// Only necessary for Go < 1.6.
		if err, ok := err.Err.(net.Error); ok {
			return err.Timeout()
		}
	}
	return false
}

func init() {
	post(handlerInit)
}

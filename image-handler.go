// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/die-net/fotomat/imager"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

var (
	maxOutputDimension      = flag.Int("max_output_dimension", 2048, "Maximum width or height of an image response.")
	maxBufferPixels         = flag.Int("max_buffer_pixels", 6500000, "Maximum number of pixels to allocate for an intermediate image buffer.")
	losslessMaxBitsPerPixel = flag.Int("lossless_max_bits_per_pixel", 4, "If saving in lossless format exceeds this size, switch to lossy (0=disable).")
	maxProcessingDuration   = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0 = disable).")
	localImageDirectory     = flag.String("local_image_directory", "", "Enable local image serving from this path (\"\" = proxy instead).")
	pool                    chan bool
	transport               = http.Transport{Proxy: http.ProxyFromEnvironment}
	client                  = http.Client{Transport: http.RoundTripper(&transport)}
)

func init() {
	http.HandleFunc("/", imageProxyHandler)
	http.HandleFunc("/albums/crop", albumsCropHandler)
}

func imageProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		sendError(w, nil, http.StatusMethodNotAllowed)
		return
	}

	path, preview, crop, width, height, ok := parsePath(r.URL.Path)
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

	fetchAndProcessImage(w, u.String(), preview, crop, width, height)
}

var matchPath = regexp.MustCompile(`^(/.*)=(p?)([sc])(\d{1,5})x(\d{1,5})$`)

func parsePath(path string) (string, bool, bool, int, int, bool) {
	g := matchPath.FindStringSubmatch(path)
	if len(g) != 6 {
		return "", false, false, 0, 0, false
	}

	// Disallow repeated scaling parameters.
	if matchPath.MatchString(g[1]) {
		return "", false, false, 0, 0, false
	}

	width, err := strconv.Atoi(g[4])
	if err != nil || width <= 0 || width > *maxOutputDimension {
		return "", false, false, 0, 0, false
	}

	height, err := strconv.Atoi(g[5])
	if err != nil || height <= 0 || height > *maxOutputDimension {
		return "", false, false, 0, 0, false
	}

	return g[1], (g[2] == "p"), (g[3] == "c"), int(width), int(height), true
}

func poolInit(limit int) {
	if *localImageDirectory != "" {
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(*localImageDirectory)))
	}

	pool = make(chan bool, limit)
	for i := 0; i < limit; i++ {
		pool <- true
	}
}

/*
        Supported geometries:
	WxH#        - scale down so the shorter edge fits within this bounding box, crop to new aspect ratio
	WxH or WxH> - scale down so the longer edge fits within this bounding box, no crop
*/
var matchGeometry = regexp.MustCompile(`^(\d{1,5})x(\d{1,5})([>#])?$`)

func albumsCropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		sendError(w, nil, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendError(w, err, 0)
		return
	}

	crop, width, height, ok := parseGeometry(r.FormValue("geometry"))
	if !ok {
		sendError(w, nil, 400)
		return
	}

	fetchAndProcessImage(w, r.FormValue("image_url"), false, crop, width, height)
}

func fetchAndProcessImage(w http.ResponseWriter, url string, preview, crop bool, width, height int) {
	aborted := w.(http.CloseNotifier).CloseNotify()

	orig, status, err := fetchURL(url)
	if err != nil || status != http.StatusOK {
		sendError(w, err, status)
		return
	}

	// Wait for an image thread to be available.
	<-pool

	// Has client closed connection while we were waiting?
	select {
	case <-aborted:
		pool <- true // Free up image thread ASAP.
		sendError(w, nil, http.StatusRequestTimeout)
		return
	default:
	}

	thumb, err := processImage(url, orig, preview, crop, width, height)
	orig = nil // Free up image memory ASAP.

	pool <- true // Free up image thread ASAP.

	if err != nil {
		thumb = nil // Free up image memory ASAP.
		sendError(w, err, 0)
		return
	}

	w.Header().Set("Server", "Fotomat")
	w.Header().Set("Content-Length", strconv.Itoa(len(thumb)))
	w.Write(thumb)
	thumb = nil // Free up image memory ASAP.
}

func parseGeometry(geometry string) (bool, int, int, bool) {
	g := matchGeometry.FindStringSubmatch(geometry)
	if len(g) != 4 {
		return false, 0, 0, false
	}
	width, err := strconv.Atoi(g[1])
	if err != nil || width <= 0 || width > *maxOutputDimension {
		return false, 0, 0, false
	}
	height, err := strconv.Atoi(g[2])
	if err != nil || height <= 0 || height > *maxOutputDimension {
		return false, 0, 0, false
	}
	crop := (g[3] == "#")
	return crop, width, height, true
}

func fetchURL(url string) ([]byte, int, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	switch resp.StatusCode {
	case http.StatusOK,
		http.StatusNoContent,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusRequestTimeout,
		http.StatusGone:
		return body, resp.StatusCode, nil
	default:
		err := fmt.Errorf("Proxy received %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return nil, http.StatusBadGateway, err
	}
}

func processImage(url string, orig []byte, preview, crop bool, width, height int) ([]byte, error) {
	if *maxProcessingDuration > 0 {
		timer := time.AfterFunc(*maxProcessingDuration, func() {
			panic(fmt.Sprintf("Processing %v longer than %v", url, *maxProcessingDuration))
		})
		defer timer.Stop()
	}

	options := imager.Options{
		Width:           width,
		Height:          height,
		MaxBufferPixels: *maxBufferPixels,
		Crop:            crop,
		Sharpen:         true,
		SaveOptions: imager.SaveOptions{
			LosslessMaxBitsPerPixel: *losslessMaxBitsPerPixel,
		},
	}

	// Preview images are tiny, blurry JPEGs.
	if preview {
		options.Sharpen = false
		options.BlurSigma = 0.8
		options.Format = imager.Jpeg
		options.Quality = 40
	}

	return imager.Thumbnail(orig, options)
}

func sendError(w http.ResponseWriter, err error, status int) {
	if status == 0 {
		switch err {
		case imager.ErrUnknownFormat, imager.ErrTooSmall:
			status = http.StatusUnsupportedMediaType
		case imager.ErrTooBig:
			status = http.StatusRequestEntityTooLarge
		default:
			status = http.StatusInternalServerError
		}
	}
	if err == nil {
		err = fmt.Errorf(http.StatusText(status))
	}
	http.Error(w, err.Error(), status)
}

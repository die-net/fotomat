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
	maxOutputDimension    = flag.Int("max_output_dimension", 2048, "Maximum width or height of an image response.")
	maxBufferPixels       = flag.Uint("max_buffer_pixels", 6500000, "Maximum number of pixels to allocate for an intermediate image buffer.")
	maxProcessingDuration = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0 = disable).")
	pool                  chan bool
	transport             http.RoundTripper = &http.Transport{Proxy: http.ProxyFromEnvironment}
	client                                  = http.Client{Transport: transport}
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

	path, crop, width, height, ok := parsePath(r.URL.Path)
	if !ok {
		sendError(w, nil, 400)
		return
	}

	u := &url.URL{Scheme: "http", Host: r.Host, Path: path}

	fetchAndProcessImage(w, u.String(), crop, width, height)
}

var matchPath = regexp.MustCompile(`^(/.*)=([sc])(\d{1,5})x(\d{1,5})$`)

func parsePath(path string) (string, bool, uint, uint, bool) {
	g := matchPath.FindStringSubmatch(path)
	if len(g) != 5 {
		return "", false, 0, 0, false
	}

	width, err := strconv.Atoi(g[3])
	if err != nil || width <= 0 || width > *maxOutputDimension {
		return "", false, 0, 0, false
	}

	height, err := strconv.Atoi(g[4])
	if err != nil || height <= 0 || height > *maxOutputDimension {
		return "", false, 0, 0, false
	}

	return g[1], (g[2] == "c"), uint(width), uint(height), true
}

func poolInit(limit int) {
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

	fetchAndProcessImage(w, r.FormValue("image_url"), crop, width, height)
}

func fetchAndProcessImage(w http.ResponseWriter, url string, crop bool, width, height uint) {
	aborted := w.(http.CloseNotifier).CloseNotify()

	orig, err, status := fetchUrl(url)
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

	thumb, err := processImage(url, orig, crop, width, height)
	orig = nil // Free up image memory ASAP.

	pool <- true // Free up image thread ASAP.

	if err != nil {
		thumb = nil // Free up image memory ASAP.
		sendError(w, err, 0)
		return
	}

	w.Write(thumb)
	thumb = nil // Free up image memory ASAP.
}

func parseGeometry(geometry string) (bool, uint, uint, bool) {
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
	return crop, uint(width), uint(height), true
}

func fetchUrl(url string) ([]byte, error, int) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err, 0
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err, 0
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
		return body, nil, resp.StatusCode
	default:
		err := fmt.Errorf("Proxy received %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return nil, err, http.StatusBadGateway
	}
}

func processImage(url string, orig []byte, crop bool, width, height uint) ([]byte, error) {
	if *maxProcessingDuration > 0 {
		timer := time.AfterFunc(*maxProcessingDuration, func() {
			panic(fmt.Sprintf("Processing %v longer than %v", url, *maxProcessingDuration))
		})
		defer timer.Stop()
	}

	img, err := imager.New(orig, *maxBufferPixels)
	if err != nil {
		return nil, err
	}

	defer img.Close()

	var thumb []byte
	if crop {
		thumb, err = img.Crop(width, height)
	} else {
		thumb, err = img.Thumbnail(width, height, true)
	}
	return thumb, err
}

func sendError(w http.ResponseWriter, err error, status int) {
	if status == 0 {
		switch err {
		case imager.UnknownFormat:
			status = http.StatusUnsupportedMediaType
		case imager.TooBig:
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

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
	"regexp"
	"strconv"
	"time"
)

var (
	maxOutputDimension    = flag.Int("max_output_dimension", 2048, "Maximum width or height of an image response.")
	maxBufferPixels       = flag.Uint("max_buffer_pixels", 6500000, "Maximum number of pixels to allocate for an intermediate image buffer.")
	maxProcessingDuration = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0 = disable).")
	pool                  chan bool
)

func init() {
	http.HandleFunc("/albums/crop", imageCropHandler)
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
var (
	matchGeometry = regexp.MustCompile(`^(\d{1,5})x(\d{1,5})([>#])?$`)
)

func imageCropHandler(w http.ResponseWriter, r *http.Request) {
	aborted := w.(http.CloseNotifier).CloseNotify()

	if r.Method != "GET" && r.Method != "HEAD" {
		sendError(w, nil, http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendError(w, err, 0)
		return
	}

	width, height, crop, ok := parseGeometry(r.FormValue("geometry"))
	if !ok {
		sendError(w, nil, 400)
		return
	}

	url := r.FormValue("image_url")
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

	thumb, err := processImage(url, orig, width, height, crop)
	orig = nil // Free up image memory ASAP.

	pool <- true // Free up image thread ASAP.

	if err != nil {
		sendError(w, err, 0)
		return
	}

	w.Write(thumb)
}

func parseGeometry(geometry string) (uint, uint, bool, bool) {
	g := matchGeometry.FindStringSubmatch(geometry)
	if len(g) != 4 {
		return 0, 0, false, false
	}
	width, err := strconv.Atoi(g[1])
	if err != nil || width <= 0 || width >= *maxOutputDimension {
		return 0, 0, false, false
	}
	height, err := strconv.Atoi(g[2])
	if err != nil || height <= 0 || height >= *maxOutputDimension {
		return 0, 0, false, false
	}
	crop := (g[3] == "#")
	return uint(width), uint(height), crop, true
}

func fetchUrl(url string) ([]byte, error, int) {
	resp, err := http.Get(url)
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

func processImage(url string, orig []byte, width, height uint, crop bool) ([]byte, error) {
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

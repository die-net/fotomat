package main

import (
	"flag"
	"fmt"
	"github.com/die-net/fotomat/imager"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

var maxBufferDimension = flag.Uint("max_buffer_dimension", 2048, "Maximum width or height of an image buffer to allocate.")

func init() {
	http.HandleFunc("/albums/crop", imageCropHandler)
}

/*
        Supported geometries:
	WxH#        - scale down so the shorter edge fits within this bounding box, crop to new aspect ratio
	WxH or WxH> - scale down so the longer edge fits within this bounding box, no crop
*/
var (
	matchGeometry = regexp.MustCompile(`^(\d{1,5})x(\d{1,5})([>#])?$`)
)

const maxDimension = 2048

func imageCropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		sendError(w, nil, http.StatusMethodNotAllowed)
	}

	if err := r.ParseForm(); err != nil {
		sendError(w, err, 0)
	}

	orig, err, status := fetchUrl(r.FormValue("image_url"))
	if err != nil || status != http.StatusOK {
		sendError(w, err, status)
		return
	}

	width, height, crop, ok := parseGeometry(r.FormValue("geometry"))
	if !ok {
		sendError(w, nil, 400)
		return
	}

	img, err := imager.New(orig, *maxBufferDimension)
	if err != nil {
		sendError(w, err, 0)
		return
	}

	defer img.Close()

	var thumb []byte
	if crop {
		thumb, err = img.Crop(width, height)
	} else {
		thumb, err = img.Thumbnail(width, height, true)
	}
	if err != nil {
		sendError(w, err, 0)
		return
	}

	w.Write(thumb)
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
	case http.StatusOK, http.StatusNoContent, http.StatusBadRequest,
		http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
		http.StatusRequestTimeout, http.StatusGone:

		return body, nil, resp.StatusCode
	default:
		err := fmt.Errorf("Proxy received %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return nil, err, http.StatusBadGateway
	}
}

func parseGeometry(geometry string) (uint, uint, bool, bool) {
	g := matchGeometry.FindStringSubmatch(geometry)
	if len(g) != 4 {
		return 0, 0, false, false
	}
	width, err := strconv.Atoi(g[1])
	if err != nil || width <= 0 || width >= maxDimension {
		return 0, 0, false, false
	}
	height, err := strconv.Atoi(g[2])
	if err != nil || height <= 0 || height >= maxDimension {
		return 0, 0, false, false
	}
	crop := (g[3] == "#")
	return uint(width), uint(height), crop, true
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

package main

import (
	"fmt"
	"github.com/die-net/fotomat/imager"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

func init() {
	http.HandleFunc("/albums/crop", imageCropHandler)
}

var (
	matchGeometry = regexp.MustCompile(`^(\d+)x(\d+)([<>!%#])?$`)
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
	if err != nil {
		sendError(w, err, status)
		return
	}

	width, height, _, ok := parseGeometry(r.FormValue("geometry"))
	if !ok {
		sendError(w, nil, 400)
		return
	}

	img, err := imager.New(orig)
	if err != nil {
		sendError(w, err, 0)
		return
	}

	defer img.Close()

	thumb, err := img.Thumbnail(width, height, true)
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
	case http.StatusOK, http.StatusNoContent, http.StatusNotFound:
		return body, nil, resp.StatusCode
	default:
		err := fmt.Errorf("Proxy received %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return nil, err, http.StatusBadGateway
	}
}

func parseGeometry(geometry string) (uint, uint, string, bool) {
	g := matchGeometry.FindStringSubmatch(geometry)
	if len(g) != 4 {
		return 0, 0, "", false
	}
	width, err := strconv.Atoi(g[1])
	if err != nil || width <= 0 || width >= maxDimension {
		return 0, 0, "", false
	}
	height, err := strconv.Atoi(g[2])
	if err != nil || height <= 0 || height >= maxDimension {
		return 0, 0, "", false
	}
	mode := g[3]
	return uint(width), uint(height), mode, true
}

func sendError(w http.ResponseWriter, err error, status int) {
	if status == 0 {
		switch err {
		case imager.UnknownFormat:
			status = http.StatusNotFound
		case imager.TooBig:
			status = http.StatusForbidden
		default:
			status = http.StatusInternalServerError
		}
	}
	if err == nil {
		err = fmt.Errorf(http.StatusText(status))
	}
	http.Error(w, err.Error(), status)
}

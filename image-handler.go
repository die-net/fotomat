package main

import (
	"fmt"
	"github.com/die-net/fotomat/imager"
	"io/ioutil"
	"net/http"
)

func init() {
	http.HandleFunc("/image/crop", imageCropHandler)
}

func imageCropHandler(w http.ResponseWriter, r *http.Request) {
	orig, err, status := fetchUrl("http://www.die.net/style/logo-solid.gif")
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	img, err := imager.New(orig)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	defer img.Close()

	thumb, err := img.Thumbnail(40, 40, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(thumb)
}

func fetchUrl(url string) ([]byte, error, int) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err, 500
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err, 500
	}

	switch resp.StatusCode {
	case 200, 204, 404:
		return body, nil, resp.StatusCode
	default:
		err := fmt.Errorf("Proxy received %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return nil, err, 502
	}
}

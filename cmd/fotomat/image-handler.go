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
	"regexp"
	"strconv"
	"time"
)

var (
	maxOutputDimension    = flag.Int("max_output_dimension", 2048, "Maximum width or height of an image response.")
	maxBufferPixels       = flag.Int("max_buffer_pixels", 6500000, "Maximum number of pixels to allocate for an intermediate image buffer.")
	sharpen               = flag.Bool("sharpen", false, "Sharpen after resize.")
	alwaysInterpolate     = flag.Bool("always_interpolate", false, "Always use slower high-quality interpolator for final 2x shrink.")
	lossless              = flag.Bool("lossless", true, "Allow saving as PNG even without transparency.")
	lossyIfPhoto          = flag.Bool("lossy_if_photo", true, "Save as lossy if image is detected as a photo.")
	losslessWebp          = flag.Bool("lossless_webp", false, "When saving in WebP, allow lossless encoding.")
	fetchTimeout          = flag.Duration("fetch_timeout", 30*time.Second, "How long to wait to receive original image from source (0=disable).")
	maxProcessingDuration = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0=disable).")
	localImageDirectory   = flag.String("local_image_directory", "", "Enable local image serving from this path (\"\" = proxy instead).")
	pool                  *thumbnail.Pool
	fetchSemaphore        chan bool
	transport             http.Transport
	client                http.Client
)

func init() {
	http.HandleFunc("/", imageProxyHandler)
}

func poolInit(workers, fetchLimit int) {
	transport = http.Transport{Proxy: http.ProxyFromEnvironment}
	if *localImageDirectory != "" {
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(*localImageDirectory)))
	}

	client = http.Client{Transport: http.RoundTripper(&transport), Timeout: *fetchTimeout}

	pool = thumbnail.NewPool(workers, 1)

	fetchSemaphore = make(chan bool, fetchLimit)
	for i := 0; i < fetchLimit; i++ {
		fetchSemaphore <- true
	}

}

func imageProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		sendError(w, nil, http.StatusMethodNotAllowed)
		return
	}

	path, options, saveOptions, ok := parsePath(r.URL.Path)
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

var matchPath = regexp.MustCompile(`^(/.*)=(p?)(w?)([sc])(\d{1,5})x(\d{1,5})$`)

func parsePath(path string) (string, thumbnail.Options, format.SaveOptions, bool) {
	g := matchPath.FindStringSubmatch(path)
	if len(g) != 7 {
		return "", thumbnail.Options{}, format.SaveOptions{}, false
	}

	path = g[1]
	preview := g[2] == "p"
	webp := g[3] == "w"
	crop := g[4] == "c"
	width, _ := strconv.Atoi(g[5])
	height, _ := strconv.Atoi(g[6])

	// Disallow repeated scaling parameters.
	if matchPath.MatchString(path) {
		return "", thumbnail.Options{}, format.SaveOptions{}, false
	}

	if width <= 0 || height <= 0 || width > *maxOutputDimension || height > *maxOutputDimension {
		return "", thumbnail.Options{}, format.SaveOptions{}, false
	}

	o := thumbnail.Options{
		Width:             width,
		Height:            height,
		MaxBufferPixels:   *maxBufferPixels,
		Sharpen:           *sharpen,
		Crop:              crop,
		AlwaysInterpolate: *alwaysInterpolate,
	}

	so := format.SaveOptions{
		Lossless:     *lossless,
		LossyIfPhoto: *lossyIfPhoto,
	}

	// Preview images are tiny, blurry JPEGs.
	if preview {
		o.Sharpen = false
		o.BlurSigma = 0.8
		so.Format = format.Jpeg
		so.Quality = 40
	}

	if webp {
		so.AllowWebp = true
		if so.Format != format.Unknown {
			so.Format = format.Webp
		}
		so.Lossless = *losslessWebp
	}

	return path, o, so, true

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

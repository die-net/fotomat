package main

import (
	"flag"
	"fmt"
	"github.com/die-net/fotomat/format"
	"github.com/die-net/fotomat/thumbnail"
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
	sharpen                 = flag.Bool("sharpen", true, "Sharpen after resize.")
	losslessMaxBitsPerPixel = flag.Int("lossless_max_bits_per_pixel", 4, "If saving in lossless format exceeds this size, switch to lossy (0=always lossy).")
	fetchTimeout            = flag.Duration("fetch_timeout", 30*time.Second, "How long to wait to receive original image from source (0=disable).")
	maxProcessingDuration   = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0=disable).")
	localImageDirectory     = flag.String("local_image_directory", "", "Enable local image serving from this path (\"\" = proxy instead).")
	pool                    chan bool
	transport               http.Transport
	client                  http.Client
)

func init() {
	http.HandleFunc("/", imageProxyHandler)
}

func poolInit(limit int) {
	transport = http.Transport{Proxy: http.ProxyFromEnvironment}
	if *localImageDirectory != "" {
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(*localImageDirectory)))
	}

	client = http.Client{Transport: http.RoundTripper(&transport), Timeout: *fetchTimeout}

	pool = make(chan bool, limit)
	for i := 0; i < limit; i++ {
		pool <- true
	}
}

func imageProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		sendError(w, nil, http.StatusMethodNotAllowed)
		return
	}

	path, preview, webp, crop, width, height, ok := parsePath(r.URL.Path)
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

	fetchAndProcessImage(w, u.String(), preview, webp, crop, width, height)
}

var matchPath = regexp.MustCompile(`^(/.*)=(p?)(w?)([sc])(\d{1,5})x(\d{1,5})$`)

func parsePath(path string) (string, bool, bool, bool, int, int, bool) {
	g := matchPath.FindStringSubmatch(path)
	if len(g) != 7 {
		return "", false, false, false, 0, 0, false
	}

	// Disallow repeated scaling parameters.
	if matchPath.MatchString(g[1]) {
		return "", false, false, false, 0, 0, false
	}

	width, err := strconv.Atoi(g[5])
	if err != nil || width <= 0 || width > *maxOutputDimension {
		return "", false, false, false, 0, 0, false
	}

	height, err := strconv.Atoi(g[6])
	if err != nil || height <= 0 || height > *maxOutputDimension {
		return "", false, false, false, 0, 0, false
	}

	return g[1], (g[2] == "p"), (g[3] == "w"), (g[4] == "c"), int(width), int(height), true
}

func fetchAndProcessImage(w http.ResponseWriter, url string, preview, webp, crop bool, width, height int) {
	aborted := w.(http.CloseNotifier).CloseNotify()

	resp, err := client.Get(url)
	if err != nil {
		sendError(w, err, 0)
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

	orig, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		pool <- true // Free up image thread ASAP.
		resp.Body.Close()
		sendError(w, err, resp.StatusCode)
		return
	}

	resp.Body.Close()

	thumb, err := processImage(url, orig, preview, webp, crop, width, height)
	orig = nil // Free up image memory ASAP.

	pool <- true // Free up image thread ASAP.

	if err != nil {
		sendError(w, err, 0)
		return
	}

	w.Header().Set("Server", "Fotomat")
	w.Header().Set("Content-Length", strconv.Itoa(len(thumb)))
	w.Write(thumb)
}

func processImage(url string, orig []byte, preview, webp, crop bool, width, height int) ([]byte, error) {
	if *maxProcessingDuration > 0 {
		timer := time.AfterFunc(*maxProcessingDuration, func() {
			panic(fmt.Sprintf("Processing %v longer than %v", url, *maxProcessingDuration))
		})
		defer timer.Stop()
	}

	options := thumbnail.Options{
		Width:           width,
		Height:          height,
		MaxBufferPixels: *maxBufferPixels,
		Crop:            crop,
		Sharpen:         *sharpen,
	}

	saveOptions := format.SaveOptions{
		LosslessMaxBitsPerPixel: *losslessMaxBitsPerPixel,
	}

	// Preview images are tiny, blurry JPEGs.
	if preview {
		options.Sharpen = false
		options.BlurSigma = 0.8
		saveOptions.Format = format.Jpeg
		saveOptions.Quality = 40
	}

	if webp {
		saveOptions.Format = format.Webp
		saveOptions.LosslessMaxBitsPerPixel = 0 // Always use lossy
	}

	return thumbnail.Thumbnail(orig, options, saveOptions)
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
			status = http.StatusInternalServerError
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

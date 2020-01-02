package main

import (
	"flag"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/die-net/fotomat/v2/format"
	"github.com/die-net/fotomat/v2/thumbnail"
)

var (
	allowPdf              = flag.Bool("allow_pdf", false, "Allow PDF as an input format")
	allowSvg              = flag.Bool("allow_svg", false, "Allow SVG as an input format")
	allowTiff             = flag.Bool("allow_tiff", false, "Allow TIFF as an input format")
	fetchTimeout          = flag.Duration("fetch_timeout", 30*time.Second, "How long to wait to receive original image from source (0=disable).")
	localImageDirectory   = flag.String("local_image_directory", "", "Enable local image serving from this path (\"\"=proxy instead).")
	lossless              = flag.Bool("lossless", true, "Allow saving as PNG even without transparency.")
	lossyIfPhoto          = flag.Bool("lossy_if_photo", true, "Save as lossy if image is detected as a photo.")
	losslessWebp          = flag.Bool("lossless_webp", false, "When saving in WebP, allow lossless encoding.")
	maxBufferPixels       = flag.Int("max_buffer_pixels", 6500000, "Maximum number of pixels to allocate for an intermediate image buffer.")
	maxImageThreads       = flag.Int("max_image_threads", numCPUCores(), "Maximum number of threads simultaneously processing images (0=all CPUs).")
	maxOutputDimension    = flag.Int("max_output_dimension", 2048, "Maximum width or height of an image response.")
	maxPrefetch           = flag.Int("max_prefetch", numCPUCores(), "Maximum number of images to prefetch before thread is available.")
	maxProcessingDuration = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0=disable).")
	maxQueueDuration      = flag.Duration("max_queue_duration", 10*time.Second, "Maximum delay of pre-image-fetch queue before returning error (0=disable).")
	sharpen               = flag.Bool("sharpen", false, "Sharpen after resize.")

	matchPath = regexp.MustCompile(`^(/.*)=(p?)(w?)([sc])(\d{1,5})x(\d{1,5})$`)
)

func handleInit() http.Handler {
	pool := thumbnail.NewPool(*maxImageThreads, 1)

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if *localImageDirectory != "" {
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(*localImageDirectory)))
	}

	client := &http.Client{Transport: http.RoundTripper(transport), Timeout: *fetchTimeout}

	return thumbnail.NewProxy(director, pool, *maxPrefetch+*maxImageThreads, client)
}

func director(req *http.Request) (thumbnail.Options, int) {
	g := matchPath.FindStringSubmatch(req.URL.Path)
	if len(g) != 7 {
		return thumbnail.Options{}, http.StatusBadRequest
	}

	if *localImageDirectory != "" {
		req.URL.Scheme = "file"
		req.URL.Host = "localhost"
	} else {
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
	}

	req.URL.Path = g[1]
	preview := g[2] == "p"
	webp := g[3] == "w"
	crop := g[4] == "c"
	width, _ := strconv.Atoi(g[5])
	height, _ := strconv.Atoi(g[6])

	// Disallow repeated scaling parameters.
	if matchPath.MatchString(req.URL.Path) {
		return thumbnail.Options{}, http.StatusBadRequest
	}

	if width <= 0 || height <= 0 || width > *maxOutputDimension || height > *maxOutputDimension {
		return thumbnail.Options{}, http.StatusBadRequest
	}

	o := thumbnail.Options{
		Width:                 width,
		Height:                height,
		MaxBufferPixels:       *maxBufferPixels,
		Sharpen:               *sharpen,
		Crop:                  crop,
		MaxQueueDuration:      *maxQueueDuration,
		MaxProcessingDuration: *maxProcessingDuration,
		AllowPdf:              *allowPdf,
		AllowSvg:              *allowSvg,
		AllowTiff:             *allowTiff,
		Save: format.SaveOptions{
			Lossless:     *lossless,
			LossyIfPhoto: *lossyIfPhoto,
		},
	}

	if webp {
		o.Save.AllowWebp = true
		o.Save.Lossless = *losslessWebp
	}

	// Preview images are tiny, blurry JPEGs/lossy WebPs.
	if preview {
		o.Sharpen = false
		o.BlurSigma = 0.4
		o.Save.Lossless = false
		o.Save.Quality = 40
	}

	return o, 0
}

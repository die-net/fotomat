package main

import (
	"flag"
	"github.com/die-net/fotomat/format"
	"github.com/die-net/fotomat/thumbnail"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	fastResize            = flag.Bool("fast_resize", false, "Allow faster resizing, at lower image quality in some cases.")
	fetchTimeout          = flag.Duration("fetch_timeout", 30*time.Second, "How long to wait to receive original image from source (0=disable).")
	localImageDirectory   = flag.String("local_image_directory", "", "Enable local image serving from this path (\"\"=proxy instead).")
	lossless              = flag.Bool("lossless", true, "Allow saving as PNG even without transparency.")
	lossyIfPhoto          = flag.Bool("lossy_if_photo", true, "Save as lossy if image is detected as a photo.")
	losslessWebp          = flag.Bool("lossless_webp", false, "When saving in WebP, allow lossless encoding.")
	maxBufferPixels       = flag.Int("max_buffer_pixels", 6500000, "Maximum number of pixels to allocate for an intermediate image buffer.")
	maxImageThreads       = flag.Int("max_image_threads", numCpuCores(), "Maximum number of threads simultaneously processing images (0=all CPUs).")
	maxOutputDimension    = flag.Int("max_output_dimension", 2048, "Maximum width or height of an image response.")
	maxPrefetch           = flag.Int("max_prefetch", numCpuCores(), "Maximum number of images to prefetch before thread is available.")
	maxProcessingDuration = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0=disable).")
	sharpen               = flag.Bool("sharpen", false, "Sharpen after resize.")

	matchPath = regexp.MustCompile(`^(/.*)=(p?)(w?)([sc])(\d{1,5})x(\d{1,5})$`)
)

func handleInit() {
	pool := thumbnail.NewPool(*maxImageThreads, 1)

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if *localImageDirectory != "" {
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir(*localImageDirectory)))
	}

	client := &http.Client{Transport: http.RoundTripper(transport), Timeout: *fetchTimeout}

	http.Handle("/", thumbnail.NewProxy(director, pool, *maxPrefetch+*maxImageThreads, client))
}

func director(req *http.Request) (thumbnail.Options, int) {
	g := matchPath.FindStringSubmatch(req.URL.Path)
	if len(g) != 7 {
		return thumbnail.Options{}, http.StatusBadRequest
	}

	if *localImageDirectory != "" {
		req.URL.Scheme = "file"
		req.URL.Host = "localhost"
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
		Width:           width,
		Height:          height,
		MaxBufferPixels: *maxBufferPixels,
		Sharpen:         *sharpen,
		Crop:            crop,
		FastResize:      *fastResize,
		Save: format.SaveOptions{
			Lossless:     *lossless,
			LossyIfPhoto: *lossyIfPhoto,
		},
	}

	// Preview images are tiny, blurry JPEGs.
	if preview {
		o.Sharpen = false
		o.BlurSigma = 0.4
		o.Save.Format = format.Jpeg
		o.Save.Quality = 40
	}

	if webp {
		o.Save.AllowWebp = true
		if o.Save.Format != format.Unknown {
			o.Save.Format = format.Webp
		}
		o.Save.Lossless = *losslessWebp
	}

	return o, 0
}

func init() {
	post(handleInit)
}

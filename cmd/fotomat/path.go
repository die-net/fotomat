package main

import (
	"flag"
	"github.com/die-net/fotomat/format"
	"github.com/die-net/fotomat/thumbnail"
	"regexp"
	"strconv"
	"time"
)

var (
	fastResize            = flag.Bool("fast_resize", false, "Allow faster resizing, at lower image quality in some cases.")
	lossless              = flag.Bool("lossless", true, "Allow saving as PNG even without transparency.")
	lossyIfPhoto          = flag.Bool("lossy_if_photo", true, "Save as lossy if image is detected as a photo.")
	losslessWebp          = flag.Bool("lossless_webp", false, "When saving in WebP, allow lossless encoding.")
	maxBufferPixels       = flag.Int("max_buffer_pixels", 6500000, "Maximum number of pixels to allocate for an intermediate image buffer.")
	maxOutputDimension    = flag.Int("max_output_dimension", 2048, "Maximum width or height of an image response.")
	maxProcessingDuration = flag.Duration("max_processing_duration", time.Minute, "Maximum duration we can be processing an image before assuming we crashed (0=disable).")
	sharpen               = flag.Bool("sharpen", false, "Sharpen after resize.")

	matchPath = regexp.MustCompile(`^(/.*)=(p?)(w?)([sc])(\d{1,5})x(\d{1,5})$`)
)

func pathParse(path string) (string, thumbnail.Options, bool) {
	g := matchPath.FindStringSubmatch(path)
	if len(g) != 7 {
		return "", thumbnail.Options{}, false
	}

	path = g[1]
	preview := g[2] == "p"
	webp := g[3] == "w"
	crop := g[4] == "c"
	width, _ := strconv.Atoi(g[5])
	height, _ := strconv.Atoi(g[6])

	// Disallow repeated scaling parameters.
	if matchPath.MatchString(path) {
		return "", thumbnail.Options{}, false
	}

	if width <= 0 || height <= 0 || width > *maxOutputDimension || height > *maxOutputDimension {
		return "", thumbnail.Options{}, false
	}

	o := thumbnail.Options{
		Width:              width,
		Height:             height,
		MaxBufferPixels:    *maxBufferPixels,
		Sharpen:            *sharpen,
		Crop:               crop,
		FastResize:         *fastResize,
		IccProfileFilename: sRgbFile,
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

	return path, o, true

}

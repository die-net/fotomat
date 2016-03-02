package format

import (
	"errors"
	"github.com/die-net/fotomat/vips"
)

const (
	DefaultQuality     = 85
	DefaultCompression = 6
)

var ErrInvalidSaveFormat = errors.New("Invalid save format")

type SaveOptions struct {
	Format       Format
	Quality      int
	Compression  int
	AllowWebp    bool
	Lossless     bool
	LossyIfPhoto bool
}

func Save(image *vips.Image, options SaveOptions) ([]byte, error) {
	if options.Quality < 1 || options.Quality > 100 {
		options.Quality = DefaultQuality
	}

	if options.Compression < 1 || options.Compression > 9 {
		options.Compression = DefaultCompression
	}

	// Make a decision on image format and whether we're using lossless.
	if options.Format == Unknown {
		if options.AllowWebp {
			options.Format = Webp
		} else if image.HasAlpha() || useLossless(image, options) {
			options.Format = Png
		} else {
			options.Format = Jpeg
		}
	} else if options.Format == Webp && !useLossless(image, options) {
		options.Lossless = false
	}

	switch options.Format {
	case Jpeg:
		return jpegSave(image, options)
	case Png:
		return pngSave(image, options)
	case Webp:
		return webpSave(image, options)
	default:
		return nil, ErrInvalidSaveFormat
	}
}

func jpegSave(image *vips.Image, options SaveOptions) ([]byte, error) {
	// JPEG interlace saves 2-3%, but incurs a few hundred bytes of
	// overhead, requires buffering the image completely in RAM for
	// encoding and decoding, and takes over 3x the CPU.  This isn't
	// usually beneficial on small images and is too expensive for large
	// images.
	pixels := image.Xsize() * image.Ysize()
	interlace := pixels >= 200*200 && pixels <= 1024*1024

	// Strip and optimize both save space, enable them.
	return image.JpegsaveBuffer(true, options.Quality, true, interlace)
}

func pngSave(image *vips.Image, options SaveOptions) ([]byte, error) {
	// PNG interlace is larger; don't use it.
	return image.PngsaveBuffer(options.Compression, false)
}

func webpSave(image *vips.Image, options SaveOptions) ([]byte, error) {
	return image.WebpsaveBuffer(options.Quality, options.Lossless)
}

func useLossless(image *vips.Image, options SaveOptions) bool {
	if !options.Lossless {
		return false
	}

	if !options.LossyIfPhoto {
		return true
	}

	// Mobile devices start being unwilling to load >= 3 megapixel PNGs.
	// Also we don't want to bother to edge detect on large images.
	if image.Xsize()*image.Ysize() >= 3*1024*1024 {
		return false
	}

	// Take a histogram of a Sobel edge detect of our image.  What's the
	// highest number of histogram values in a row that are more than 1%
	// of the maximum value? Above 16 indicates a photo.
	metric, err := image.PhotoMetric(0.01)
	return err != nil || metric < 16
}

package format

import (
	"errors"
	"github.com/die-net/fotomat/vips"
)

const (
	// DefaultQuality is used when SaveOptions.Quality is unspecified.
	DefaultQuality = 85
	// DefaultCompression is used when SaveOptions.Compression is unspecified.
	DefaultCompression = 6
)

// ErrInvalidSaveFormat is returned if the specified Format can't be written to.
var ErrInvalidSaveFormat = errors.New("Invalid save format")

// SaveOptions specifies how an image should be saved.
type SaveOptions struct {
	// Format is the Format that an image is saved in. If unspecified, the best output format for a given input image is selected.
	Format Format
	// JPEG or WebP quality for an output image (1-100).
	Quality int
	// Compress is the GZIP compression setting to use for PNG images (1-9).
	Compression int
	// AllowWebp allows automatic selection of WebP format, if reader can support it.
	AllowWebp bool
	// Lossless allows selection of a lossless output format.
	Lossless bool
	// LossyIfPhoto uses a lossy format if it detects that an image is a photo.
	LossyIfPhoto bool
}

// Save returns an Image compressed using the given SaveOptions as a byte slice.
func Save(image *vips.Image, options SaveOptions) ([]byte, error) {
	if options.Quality < 1 || options.Quality > 100 {
		options.Quality = DefaultQuality
	}

	if options.Compression < 1 || options.Compression > 9 {
		options.Compression = DefaultCompression
	}

	// Make a decision on image format and whether we're using lossless.
	if options.Format == Unknown {
		switch {
		case options.AllowWebp:
			options.Format = Webp
		case image.HasAlpha() || useLossless(image, options):
			options.Format = Png
		default:
			options.Format = Jpeg
		}
	}

	switch options.Format {
	case Jpeg:
		return jpegSave(image, options)
	case Png:
		return pngSave(image, options)
	case Webp:
		options.Lossless = useLossless(image, options)
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
	return image.PngsaveBuffer(true, options.Compression, false)
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

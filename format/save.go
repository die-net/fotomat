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
	Format                  Format
	Quality                 int
	Compression             int
	LosslessMaxBitsPerPixel int
}

func Save(image *vips.Image, options SaveOptions) ([]byte, error) {
	if options.Quality < 1 || options.Quality > 100 {
		options.Quality = DefaultQuality
	}

	if options.Compression < 1 || options.Compression > 9 {
		options.Compression = DefaultCompression
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
	blob, err := image.PngsaveBuffer(options.Compression, false)
	if err != nil {
		return nil, err
	}

	// If PNG has transparency, keep it.
	if image.HasAlpha() {
		return blob, nil
	}

	// If LosslessMaxBitsPerPixel is suppled, and the image is not
	// larger than that, keep it.
	if options.LosslessMaxBitsPerPixel <= 0 || (len(blob)-256)*8 <= image.Xsize()*image.Ysize()*options.LosslessMaxBitsPerPixel {
		return blob, nil
	}

	// Return a Jpeg instead.
	return jpegSave(image, options)
}

func webpSave(image *vips.Image, options SaveOptions) ([]byte, error) {
	// Shall we try using lossless?
	if options.LosslessMaxBitsPerPixel > 0 {
		blob, err := image.WebpsaveBuffer(options.Quality, true)
		if err != nil {
			return nil, err
		}

		// If the image is not more than PngMaxBitsPerPixel, keep it.
		if (len(blob)-256)*8 <= image.Xsize()*image.Ysize()*options.LosslessMaxBitsPerPixel {
			return blob, nil
		}
	}

	return image.WebpsaveBuffer(options.Quality, false)
}

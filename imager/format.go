// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"errors"
	"github.com/die-net/fotomat/vips"
	"net/http"
)

var (
	ErrInvalidOperation = errors.New("Invalid Operation")
)

const (
	DefaultQuality     = 85
	DefaultCompression = 6
)

type Format int

const (
	UnknownFormat Format = iota
	Gif
	Jpeg
	Png
	WebpLossy
	WebpLossless
)

var formatInfo = []struct {
	mime string
	load func([]byte) (*vips.Image, error)
	save func(*vips.Image, Options) ([]byte, error)
}{
	{mime: "application/octet-stream", load: nil, save: nil},
	{mime: "image/gif", load: vips.MagickloadBuffer, save: nil},
	{mime: "image/jpeg", load: vips.JpegloadBuffer, save: jpegSave},
	{mime: "image/png", load: vips.PngloadBuffer, save: pngSave},
	{mime: "image/webp", load: vips.WebploadBuffer, save: webpSaveLossy},
	{mime: "image/webp", load: vips.WebploadBuffer, save: webpSaveLossless},
}

func DetectFormat(blob []byte) Format {
	mime := http.DetectContentType(blob)

	for format, info := range formatInfo {
		if info.mime == mime {
			return Format(format)
		}
	}

	return UnknownFormat
}

func (format Format) String() string {
	return formatInfo[format].mime
}

func (format Format) CanLoad() bool {
	return formatInfo[format].load != nil
}

func (format Format) Load(blob []byte) (*vips.Image, error) {
	load := formatInfo[format].load
	if load == nil {
		return nil, ErrInvalidOperation
	}
	return load(blob)
}

func (format Format) CanSave() bool {
	return formatInfo[format].save != nil
}

func (format Format) Save(image *vips.Image, options Options) ([]byte, error) {
	save := formatInfo[format].save
	if save == nil {
		return nil, ErrInvalidOperation
	}
	return save(image, options)
}

func jpegSave(image *vips.Image, options Options) ([]byte, error) {
	q := options.Quality
	if q == 0 {
		q = DefaultQuality
	}

	// JPEG interlace saves 2-3%, but incurs a few hundred bytes of
	// overhead.  This isn't usually beneficial on small images.
	interlace := image.Xsize()*image.Ysize() >= 200*200

	// Strip and optimize both save space, enable them.
	return image.JpegsaveBuffer(true, q, true, interlace)
}

func pngSave(image *vips.Image, options Options) ([]byte, error) {
	compression := options.Compression
	if compression == 0 {
		compression = DefaultCompression
	}

	// PNG interlace is larger; don't use it.
	blob, err := image.PngsaveBuffer(compression, false)
	if err != nil {
		return nil, err
	}

	// TODO: If PNG has transparency, return it.
	if false {
		return blob, nil
	}

	// If the image is larger than PngMaxBitsPerPixel, re-save as JPEG.
	if options.LosslessMaxBitsPerPixel > 0 && (len(blob)-256)*8 > image.Xsize()*image.Ysize()*options.LosslessMaxBitsPerPixel {
		return jpegSave(image, options)
	}

	return blob, nil
}

func webpSaveLossless(image *vips.Image, options Options) ([]byte, error) {
	q := options.Quality
	if q == 0 {
		q = DefaultQuality
	}

	return image.WebpsaveBuffer(q, true)
}

func webpSaveLossy(image *vips.Image, options Options) ([]byte, error) {
	q := options.Quality
	if q == 0 {
		q = DefaultQuality
	}

	return image.WebpsaveBuffer(q, false)
}

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

type Metadata struct {
	Width       int
	Height      int
	Format      Format
	Orientation Orientation
}

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
	mime      string
	loadFile  func(filename string) (*vips.Image, error)
	loadBytes func([]byte) (*vips.Image, error)
	save      func(*vips.Image, Options) ([]byte, error)
}{
	{mime: "application/octet-stream", loadFile: nil, loadBytes: nil, save: nil},
	{mime: "image/gif", loadFile: vips.Magickload, loadBytes: vips.MagickloadBuffer, save: nil},
	{mime: "image/jpeg", loadFile: vips.Jpegload, loadBytes: vips.JpegloadBuffer, save: jpegSave},
	{mime: "image/png", loadFile: vips.Pngload, loadBytes: vips.PngloadBuffer, save: pngSave},
	{mime: "image/webp", loadFile: vips.Webpload, loadBytes: vips.WebploadBuffer, save: webpSaveLossy},
	{mime: "image/webp", loadFile: vips.Webpload, loadBytes: vips.WebploadBuffer, save: webpSaveLossless},
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

func (format Format) CanLoadFile() bool {
	return formatInfo[format].loadFile != nil
}

func (format Format) CanLoadBytes() bool {
	return formatInfo[format].loadBytes != nil
}

func (format Format) MetadataFile(filename string) (Metadata, error) {
	image, err := format.LoadFile(filename)
	if err != nil {
		return Metadata{}, err
	}

	defer image.Close()

	return metadataFormat(image, format), nil
}

func MetadataBytes(blob []byte) (Metadata, error) {
	format := DetectFormat(blob)
	if format == UnknownFormat {
		return Metadata{}, ErrUnknownFormat
	}

	return format.MetadataBytes(blob)
}

func (format Format) MetadataBytes(blob []byte) (Metadata, error) {
	image, err := format.LoadBytes(blob)
	if err != nil {
		return Metadata{}, ErrUnknownFormat
	}

	defer image.Close()

	return metadataFormat(image, format), nil
}

func metadataFormat(image *vips.Image, format Format) Metadata {
	m := MetadataImage(image)
	m.Format = format
	return m
}

func MetadataImage(image *vips.Image) Metadata {
	o := DetectOrientation(image)
	w, h := o.Dimensions(image.Xsize(), image.Ysize())
	if w <= 0 || h <= 0 {
		panic("Invalid image dimensions.")
	}
	return Metadata{Width: w, Height: h, Orientation: o}
}

func (format Format) LoadFile(filename string) (*vips.Image, error) {
	loadFile := formatInfo[format].loadFile
	if loadFile == nil {
		return nil, ErrInvalidOperation
	}

	return loadFile(filename)
}

func (format Format) LoadBytes(blob []byte) (*vips.Image, error) {
	loadBytes := formatInfo[format].loadBytes
	if loadBytes == nil {
		return nil, ErrInvalidOperation
	}

	return loadBytes(blob)
}

func (format Format) CanSave() bool {
	return formatInfo[format].save != nil
}

func (format Format) Save(image *vips.Image, options Options) ([]byte, error) {
	save := formatInfo[format].save
	if save == nil {
		return nil, ErrInvalidOperation
	}

	if options.Quality == 0 {
		options.Quality = DefaultQuality
	}
	if options.Quality < 1 || options.Quality > 100 {
		return nil, ErrBadOption
	}

	if options.Compression == 0 {
		options.Compression = DefaultCompression
	}
	if options.Compression < 1 || options.Compression > 9 {
		return nil, ErrBadOption
	}

	return save(image, options)
}

func jpegSave(image *vips.Image, options Options) ([]byte, error) {
	// JPEG interlace saves 2-3%, but incurs a few hundred bytes of
	// overhead.  This isn't usually beneficial on small images.
	interlace := image.Xsize()*image.Ysize() >= 200*200

	// Strip and optimize both save space, enable them.
	return image.JpegsaveBuffer(true, options.Quality, true, interlace)
}

func pngSave(image *vips.Image, options Options) ([]byte, error) {
	// PNG interlace is larger; don't use it.
	blob, err := image.PngsaveBuffer(options.Compression, false)
	if err != nil {
		return nil, err
	}

	// If PNG has transparency, return it.
	if image.HasAlpha() {
		return blob, nil
	}

	// If the image is larger than PngMaxBitsPerPixel, re-save as JPEG.
	if options.LosslessMaxBitsPerPixel > 0 && (len(blob)-256)*8 > image.Xsize()*image.Ysize()*options.LosslessMaxBitsPerPixel {
		return jpegSave(image, options)
	}

	return blob, nil
}

func webpSaveLossless(image *vips.Image, options Options) ([]byte, error) {
	blob, err := image.WebpsaveBuffer(options.Quality, true)
	if err != nil {
		return nil, err
	}

	if options.LosslessMaxBitsPerPixel > 0 && (len(blob)-256)*8 > image.Xsize()*image.Ysize()*options.LosslessMaxBitsPerPixel {
		return webpSaveLossy(image, options)
	}

	return blob, nil
}

func webpSaveLossy(image *vips.Image, options Options) ([]byte, error) {

	return image.WebpsaveBuffer(options.Quality, false)
}

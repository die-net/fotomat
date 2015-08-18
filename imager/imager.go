// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"errors"
)

var (
	UnknownFormat = errors.New("Unknown image format")
	TooBig        = errors.New("Image is too wide or tall")
)

const (
	minDimension = 2             // Avoid off-by-one divide-by-zero errors.
	maxDimension = (2 << 15) - 2 // Avoid signed int16 overflows.
)

type Imager struct {
	blob               []byte
	Width              uint
	Height             uint
	Orientation        *Orientation
	InputFormat        string
	OutputFormat       string
	JpegQuality        uint
	PngMaxBitsPerPixel uint
	Sharpen            bool
	BlurSigma          float64
	AutoContrast       bool
}

func New(blob []byte, maxBufferPixels uint) (*Imager, error) {
	// Security: Guess at formats.  Limit formats we pass to ImageMagick
	// to just JPEG, PNG, GIF.
	inputFormat, outputFormat := detectFormats(blob)
	if inputFormat == "" {
		return nil, UnknownFormat
	}

	// Ask ImageMagick to parse metadata.
	width, height, orientation, format, err := imageMetaData(blob)
	if err != nil {
		return nil, UnknownFormat
	}

	// Assume JPEG decoder can pre-scale to 1/8 original size.
	if format == "JPEG" {
		maxBufferPixels *= 8
	}

	// Security: Confirm that detectFormat() and imageMagick agreed on
	// format and that image sizes are sane.
	if format != inputFormat {
		return nil, UnknownFormat
	} else if width < minDimension || height < minDimension {
		return nil, UnknownFormat
	} else if width > maxDimension || height > maxDimension {
		return nil, TooBig
	} else if width*height > maxBufferPixels {
		return nil, TooBig
	}

	img := &Imager{
		blob:               blob,
		Width:              width,
		Height:             height,
		Orientation:        orientation,
		InputFormat:        inputFormat,
		OutputFormat:       outputFormat,
		JpegQuality:        85,
		PngMaxBitsPerPixel: 4,
		Sharpen:            true,
		BlurSigma:          0.0,
		AutoContrast:       false,
	}

	return img, nil
}

func (img *Imager) Thumbnail(width, height uint, within bool) ([]byte, error) {
	width, height = scaleAspect(img.Width, img.Height, width, height, within)

	result, err := img.NewResult(width, height)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	if result.Width > width || result.Height > height {
		if err := result.Resize(width, height); err != nil {
			return nil, err
		}
	}

	return result.Get()
}

func (img *Imager) Crop(width, height uint) ([]byte, error) {
	// If requested width or height are larger than original, scale
	// request down to fit within original dimensions.
	if width > img.Width || height > img.Height {
		width, height = scaleAspect(width, height, img.Width, img.Height, true)
	}

	// Figure out the intermediate size the original image would have to
	// be scaled to be cropped to requested size.
	iw, ih := scaleAspect(img.Width, img.Height, width, height, false)

	result, err := img.NewResult(iw, ih)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	// If necessary, scale down to appropriate intermediate size.
	if result.Width > iw || result.Height > ih {
		if err := result.Resize(iw, ih); err != nil {
			return nil, err
		}
	}

	// If necessary, crop to fit exact size.
	if result.Width > width || result.Height > height {
		if err := result.Crop(width, height); err != nil {
			return nil, err
		}
	}

	return result.Get()
}

func (img *Imager) Close() {
	*img = Imager{}
}

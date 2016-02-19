// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"errors"
	"github.com/die-net/fotomat/vips"
	"math"
	"runtime"
)

var (
	ErrUnknownFormat = errors.New("Unknown image format")
	ErrTooBig        = errors.New("Image is too wide or tall")
	ErrTooSmall      = errors.New("Image is too small")
	ErrBadOption     = errors.New("Bad option specified")
)

const (
	minDimension = 2             // Avoid off-by-one divide-by-zero errors.
	maxDimension = (1 << 15) - 2 // Avoid signed int16 overflows.
)

func Thumbnail(blob []byte, o Options) ([]byte, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Free some thread-local caches. Safe to call unnecessarily.
	defer vips.ThreadShutdown()

	m, err := MetadataBytes(blob)
	if err != nil {
		return nil, err
	}

	if err := o.Check(m); err != nil {
		return nil, err
	}

	w, h := o.Width, o.Height

	// If requested crop width or height are larger than original, scale
	// request down to fit within original dimensions.
	if o.Crop && (w > m.Width || h > m.Height) {
		w, h = scaleAspect(w, h, m.Width, m.Height, true)
	}

	// Figure out size to scale image down to.  For crop, this is the
	// intermediate size the original image would have to be scaled to
	// be cropped to requested size.
	iw, ih := scaleAspect(m.Width, m.Height, w, h, !o.Crop)

	shrink := math.Sqrt(float64(m.Width*m.Height) / float64(iw*ih))

	// Are we shrinking by more than 2.5%?
	shrank := shrink > 1.025

	var image *vips.Image
	if m.Format == Jpeg {
		image, err = vips.JpegloadBufferShrink(blob, jpegShrink(int(shrink)))
	} else {
		image, err = m.Format.LoadBytes(blob)
	}
	if err != nil {
		return nil, err
	}

	defer image.Close()

	m = MetadataImage(image)
	if iw < m.Width || ih < m.Height {
		factor := math.Sqrt(float64(iw*ih) / float64(m.Width*m.Height))
		out, err := image.Resize(float64(factor))
		if err != nil {
			return nil, err
		}

		defer out.Close()

		image = out
		m = MetadataImage(image)
	}

	// If necessary, crop to fit exact size.
	if o.Crop && (m.Width > w || m.Height > h) {
		// Center horizontally
		x := (m.Width - w + 1) / 2
		// Assume faces are higher up vertically
		y := (m.Height - h + 1) / 4

		out, err := image.ExtractArea(m.Orientation.Crop(w, h, x, y, m.Width, m.Height))
		if err != nil {
			return nil, err
		}

		defer out.Close()

		image = out
		m = MetadataImage(image)
	}

	out, err := m.Orientation.Apply(image)
	if err != nil {
		return nil, err
	}
	if out != nil {
		defer out.Close()
		image = out
		m = MetadataImage(image)
	}

	if o.BlurSigma > 0.0 {
		out, err := image.Gaussblur(o.BlurSigma)
		if err != nil {
			return nil, err
		}

		defer out.Close()

		image = out
		m = MetadataImage(image)
	}

	if o.Sharpen && shrank {
		out, err := image.MildSharpen()
		if err != nil {
			return nil, err
		}

		defer out.Close()

		image = out
		m = MetadataImage(image)
	}

	return o.Format.Save(image, o.SaveOptions)
}

func jpegShrink(shrink int) int {
	switch {
	case shrink >= 8:
		return 8
	case shrink >= 4:
		return 4
	case shrink >= 2:
		return 2
	default:
		return 1
	}
}

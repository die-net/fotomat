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

func init() {
	vips.Initialize()
}

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

	// Figure out size to scale image down to.  For crop, this is the
	// intermediate size the original image would have to be scaled to
	// be cropped to requested size.
	iw, ih := scaleAspect(m.Width, m.Height, o.Width, o.Height, !o.Crop)

	shrink := scaleFactor(m.Width, m.Height, iw, ih)

	// Are we shrinking by more than 2.5%?
	shrank := shrink > 1.025

	image, err := load(blob, m.Format, int(shrink))
	if err != nil {
		return nil, err
	}

	image, err = preProcess(image)
	if err != nil {
		return nil, err
	}

	m = MetadataImage(image)

	// A box filter will quickly get us within 2x of the final size.
	shrink = math.Floor(scaleFactor(m.Width, m.Height, iw, ih))
	if shrink >= 2 {
		out, err := image.Shrink(shrink, shrink)
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
		m = MetadataImage(image)
	}

	// Do a high-quality resize to scale to final size.
	if iw < m.Width || ih < m.Height {
		factor := scaleFactor(iw, ih, m.Width, m.Height)
		out, err := image.Resize(float64(factor))
		if err != nil {
			return nil, err
		}

		image.Close()
		image = out
		m = MetadataImage(image)
	}

	// If necessary, crop to fit exact size.
	if o.Crop && (o.Width < m.Width || o.Height < m.Height) {
		// Center horizontally
		x := (m.Width - o.Width + 1) / 2
		// Assume faces are higher up vertically
		y := (m.Height - o.Height + 1) / 4

		out, err := image.ExtractArea(m.Orientation.Crop(o.Width, o.Height, x, y, m.Width, m.Height))
		if err != nil {
			return nil, err
		}

		image.Close()
		image = out
		m = MetadataImage(image)
	}

	image, err = postProcess(image, m.Orientation, shrank, o)
	if err != nil {
		return nil, err
	}

	thumb, err := o.Format.Save(image, o.SaveOptions)
	image.Close()
	return thumb, err
}

func scaleFactor(iw, ih, mw, mh int) float64 {
	sw := float64(iw) / float64(mw)
	if sh := float64(ih) / float64(mh); sh > sw {
		return sh
	}
	return sw
}

func load(blob []byte, format Format, shrink int) (*vips.Image, error) {
	if format == Jpeg && shrink > 1 {
		return vips.JpegloadBufferShrink(blob, jpegShrink(shrink))
	}

	return format.LoadBytes(blob)
}

func preProcess(image *vips.Image) (*vips.Image, error) {
	if out, err := image.IccImport(); err == nil {
		image.Close()
		image = out
	}

	if image.ImageGuessInterpretation() != vips.InterpretationSRGB {
		out, err := image.Colourspace(vips.InterpretationSRGB)
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
	}

	if image.HasAlpha() {
		out, err := image.Premultiply()
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
	}

	return image, nil
}

func postProcess(image *vips.Image, orientation Orientation, shrank bool, options Options) (*vips.Image, error) {
	if options.BlurSigma > 0.0 {
		out, err := image.Gaussblur(options.BlurSigma)
		if err != nil {
			return nil, err
		}

		image.Close()
		image = out
	}

	if options.Sharpen && shrank {
		out, err := image.MildSharpen()
		if err != nil {
			return nil, err
		}

		image.Close()
		image = out
	}

	if image.HasAlpha() {
		out, err := image.Unpremultiply()
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
	}

	if image.ImageGetBandFormat() != vips.BandFormatUchar {
		out, err := image.Cast(vips.BandFormatUchar)
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
	}

	out, err := orientation.Apply(image)
	if err != nil {
		return nil, err
	}
	if out != nil {
		image.Close()
		image = out
	}

	return image, nil
}

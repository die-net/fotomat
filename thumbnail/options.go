package thumbnail

import (
	"errors"
	"github.com/die-net/fotomat/format"
	"time"
)

var (
	ErrBadOption = errors.New("Bad option specified")
	ErrTooBig    = errors.New("Image is too wide or tall")
	ErrTooSmall  = errors.New("Image is too small")
)

const (
	minDimension = 2             // Avoid off-by-one divide-by-zero errors.
	maxDimension = (1 << 15) - 2 // Avoid signed int16 overflows.
)

type Options struct {
	Width                 int
	Height                int
	Crop                  bool
	MaxBufferPixels       int
	Sharpen               bool
	BlurSigma             float64
	AutoContrast          bool
	AlwaysInterpolate     bool
	MaxProcessingDuration time.Duration
}

func (o Options) Check(m format.Metadata) (Options, error) {
	// Input format must be set.
	if m.Format == format.Unknown {
		return Options{}, format.ErrUnknownFormat
	}

	// Security: Confirm that image sizes are sane.
	if m.Width < minDimension || m.Height < minDimension {
		return Options{}, ErrTooSmall
	}
	if m.Width > maxDimension || m.Height > maxDimension {
		return Options{}, ErrTooBig
	}

	// If output width or height are not set, use original.
	if o.Width == 0 {
		o.Width = m.Width
	}
	if o.Height == 0 {
		o.Height = m.Height
	}
	// Security: Verify requested width and height.
	if o.Width < 1 || o.Height < 1 {
		return Options{}, ErrTooSmall
	}
	if o.Width > maxDimension || o.Height > maxDimension {
		return Options{}, ErrTooBig
	}
	// If requested crop width or height are larger than original, scale
	// request down to fit within original dimensions.
	if o.Crop && (o.Width > m.Width || o.Height > m.Height) {
		o.Width, o.Height, _ = scaleAspect(o.Width, o.Height, m.Width, m.Height, true)
	}

	// If set, limit allocated pixels to MaxBufferPixels.  Assume JPEG
	// decoder can pre-scale to 1/8 original width and height.
	scale := 1
	if m.Format == format.Jpeg {
		scale = 8
	}
	if o.MaxBufferPixels > 0 && m.Width*m.Height > o.MaxBufferPixels*scale*scale {
		return Options{}, ErrTooBig
	}

	if o.BlurSigma < 0.0 || o.BlurSigma > 8.0 {
		return Options{}, ErrBadOption
	}

	return o, nil
}

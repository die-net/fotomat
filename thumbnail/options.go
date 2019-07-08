package thumbnail

import (
	"errors"
	"time"

	"github.com/die-net/fotomat/v2/format"
)

var (
	// ErrBadOption is returned when option values are out of range.
	ErrBadOption = errors.New("bad option specified")
	// ErrTooBig is returned when an image is too wide or tall.
	ErrTooBig = errors.New("image is too wide or tall")
	// ErrTooSmall is returned when an image is too small.
	ErrTooSmall = errors.New("image is too small")
)

const (
	minDimension = 2             // Avoid off-by-one divide-by-zero errors.
	maxDimension = (1 << 15) - 2 // Avoid signed int16 overflows.
)

// Options specifies how a Thumbnail operation should modify an image.
type Options struct {
	// Width and Height are the optional maximum sizes of output image,
	// in pixels.  If Crop is false, the original aspect ratio is
	// preserved and the more restrictive of Width or Height are used.
	Width  int
	Height int
	// Crop enables crop mode, where exact supplied Width:Height aspect
	// ratio is preserved and excess pixels are trimmed from the sides.
	Crop bool
	// Sharpen runs a mild sharpening pass on downsampled images.
	Sharpen bool
	// BlurSigma performs a gaussian blur with specified sigma.
	BlurSigma float64
	// MaxBufferPixels specifies how large of an intermediate image
	// buffer to allow, in pixels. RAM usage will be a few bytes per pixel.
	MaxBufferPixels int
	// MaxQueueDuration limits the amount of time spent in a queue before processing starts.
	MaxQueueDuration time.Duration
	// MaxProcessingDuration limits the amount of time processing an
	// image, after which it is assumed the operation has crashed and
	// the server aborts, killing all outstanding requests.
	MaxProcessingDuration time.Duration
	// Save specifies the format.SaveOptions to use when compressing the modified image.
	Save format.SaveOptions
}

// Check verifies Options against Metadata and returns a modified
// Options or an error.
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

	// If set, limit allocated pixels to MaxBufferPixels.  Assume JPEG,
	// Webp, Pdf, and Svg decoders can pre-scale to 1/8 original width and
	// height.
	scale := 1
	if m.Format == format.Jpeg || m.Format == format.Webp || m.Format == format.Pdf || m.Format == format.Svg {
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

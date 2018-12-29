package thumbnail

import (
	"fmt"
	"math"
	"time"

	"github.com/die-net/fotomat/format"
	"github.com/die-net/fotomat/vips"
)

const (
	fastResizeLimit = 1.4 // Do the last 1.4x image resize at high-quality
)

// Thumbnail scales or crops a compressed image blob according to the
// Options specified in o and returns a compressed image.
// Should be called from a thread pool with runtime.LockOSThread() locked.
func Thumbnail(blob []byte, o Options) ([]byte, error) {
	if o.MaxProcessingDuration > 0 {
		timer := time.AfterFunc(o.MaxProcessingDuration, func() {
			panic(fmt.Sprintf("Thumbnail took longer than %v", o.MaxProcessingDuration))
		})
		defer timer.Stop()
	}

	// Free some thread-local caches. Safe to call unnecessarily.
	defer vips.ThreadShutdown()

	m, err := format.MetadataBytes(blob)
	if err != nil {
		return nil, err
	}

	o, err = o.Check(m)
	if err != nil {
		return nil, err
	}

	// If source image is lossy, disable lossless.
	if m.Format == format.Jpeg {
		o.Save.Lossless = false
	}

	// Figure out size to scale image down to.  For crop, this is the
	// intermediate size the original image would have to be scaled to
	// be cropped to requested size.
	iw, ih, trustWidth := scaleAspect(m.Width, m.Height, o.Width, o.Height, !o.Crop)

	// Are we shrinking by more than 2.5%?
	shrinking := iw < m.Width-m.Width/40 && ih < m.Height-m.Height/40

	// Figure out the jpeg/webp shrink factor and load image.
	// Jpeg shrink rounds up the number of pixels.
	psf := preShrinkFactor(m.Width, m.Height, iw, ih, trustWidth, m.Format == format.Jpeg)
	image, err := load(blob, m.Format, psf)
	if err != nil {
		return nil, err
	}
	defer image.Close()

	if err := srgb(image); err != nil {
		return nil, err
	}

	if err := resize(image, iw, ih, o.BlurSigma, o.Sharpen && shrinking); err != nil {
		return nil, err
	}

	// Make sure we generate images with 8 bits per channel.  Do this before the
	// rotate to reduce the amount of data that needs to be copied.
	if image.ImageGetBandFormat() != vips.BandFormatUchar {
		if err := image.Cast(vips.BandFormatUchar); err != nil {
			return nil, err
		}
	}

	if o.Crop {
		if err := crop(image, o.Width, o.Height); err != nil {
			return nil, err
		}
	}

	if image.HasAlpha() {
		if min, err := minTransparency(image); err == nil && min >= 0.9 {
			if err := image.Flatten(); err != nil {
				return nil, err
			}
		}
	}

	if err := m.Orientation.Apply(image); err != nil {
		return nil, err
	}

	return format.Save(image, o.Save)
}

func load(blob []byte, f format.Format, shrink int) (*vips.Image, error) {
	if shrink > 1 {
		if f == format.Jpeg {
			return vips.JpegloadBufferShrink(blob, shrink)
		} else if f == format.Webp {
			return vips.WebploadBufferShrink(blob, shrink)
		} else if f == format.Pdf {
			return vips.PdfloadBufferShrink(blob, shrink)
		} else if f == format.Svg {
			return vips.SvgloadBufferShrink(blob, shrink)
		}
	}

	return f.LoadBytes(blob)
}

func srgb(image *vips.Image) error {
	// Transform from embedded ICC profile if present or default profile
	// if CMYK.  Ignore errors.
	if image.ImageFieldExists(vips.MetaIccName) {
		_ = image.IccTransform(sRgbFile, "", vips.IntentRelative)
	} else if image.ImageGuessInterpretation() == vips.InterpretationCMYK {
		_ = image.IccTransform(sRgbFile, cmykFile, vips.IntentRelative)
	}

	space := image.ImageGuessInterpretation()
	if space != vips.InterpretationSRGB && space != vips.InterpretationBW {
		if err := image.Colourspace(vips.InterpretationSRGB); err != nil {
			return err
		}
	}

	return nil
}

func resize(image *vips.Image, iw, ih int, blurSigma float64, sharpen bool) error {
	m := format.MetadataImage(image)

	// Interpolation of RGB values with an alpha channel isn't safe
	// unless the values are pre-multiplied. Undo this later.
	// This also flattens fully transparent pixels to black.
	premultiply := image.HasAlpha()
	if premultiply {
		if err := image.Premultiply(); err != nil {
			return err
		}
	}

	// Shrink is a a box filter will quickly cut the image size by
	// integer multiples, at some quality cost.
	wshrink := float64(m.Width) / (float64(iw) * fastResizeLimit)
	hshrink := float64(m.Height) / (float64(ih) * fastResizeLimit)
	shrink := math.Floor(math.Min(wshrink, hshrink))
	if shrink >= 2 {
		// Shrink rounds down the number of pixels.
		if err := image.Shrink(shrink, shrink); err != nil {
			return err
		}
		m = format.MetadataImage(image)
	}

	// If necessary, do a high-quality resize to scale to final size.
	if iw < m.Width || ih < m.Height {
		if err := image.Resize(float64(iw)/float64(m.Width), float64(ih)/float64(m.Height)); err != nil {
			return err
		}
	}

	if blurSigma > 0.0 {
		if err := image.Gaussblur(blurSigma); err != nil {
			return err
		}
	}

	if sharpen {
		if err := image.MildSharpen(); err != nil {
			return err
		}
	}

	// Unpremultiply after all operations that touch adjacent pixels.
	if premultiply {
		if err := image.Unpremultiply(); err != nil {
			return err
		}
	}

	return nil
}

func crop(image *vips.Image, ow, oh int) error {
	m := format.MetadataImage(image)

	// If we have nothing to do, return.
	if ow == m.Width && oh == m.Height {
		return nil
	}

	// Center horizontally
	x := (m.Width - ow + 1) / 2
	// Assume faces are higher up vertically
	y := (m.Height - oh + 1) / 4

	if x < 0 || y < 0 {
		panic("Bad crop offsets!")
	}

	return image.ExtractArea(m.Orientation.Crop(ow, oh, x, y, m.Width, m.Height))
}

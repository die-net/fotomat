package thumbnail

import (
	"github.com/die-net/fotomat/format"
	"github.com/die-net/fotomat/vips"
	"math"
	"runtime"
)

func Thumbnail(blob []byte, o Options, saveOptions format.SaveOptions) ([]byte, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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

	// If output format is not set, pick one.
	if saveOptions.Format == format.Unknown {
		saveOptions.Format = m.Format.SaveAs(saveOptions.LosslessMaxBitsPerPixel > 0)
	}

	// Figure out size to scale image down to.  For crop, this is the
	// intermediate size the original image would have to be scaled to
	// be cropped to requested size.
	iw, ih, trustWidth := scaleAspect(m.Width, m.Height, o.Width, o.Height, !o.Crop)

	// Are we shrinking by more than 2.5%?
	shrank := iw < m.Width-m.Width/40 && ih < m.Height-m.Height/40

	// Figure out the jpeg shrink factor and load image.
	// Jpeg shrink rounds up the number of pixels.
	js := jpegShrink(m.Width, m.Height, iw, ih, trustWidth, o.AlwaysInterpolate)

	image, err := load(blob, m.Format, js)
	if err != nil {
		return nil, err
	}
	defer image.Close()

	if err = preProcess(image); err != nil {
		return nil, err
	}

	m = format.MetadataImage(image)

	// A box filter will quickly get us within 2x of the final size.
	// Shrink rounds down the number of pixels.
	if !o.AlwaysInterpolate {
		xshrink := math.Floor(float64(m.Width) / float64(iw))
		yshrink := math.Floor(float64(m.Height) / float64(ih))
		if xshrink >= 2 || yshrink >= 2 {
			if err := image.Shrink(xshrink, yshrink); err != nil {
				return nil, err
			}
			m = format.MetadataImage(image)
		}
	}

	// Do a high-quality resize to scale to final size.
	if iw < m.Width || ih < m.Height {
		if err := image.Resize(float64(iw)/float64(m.Width), float64(ih)/float64(m.Height)); err != nil {
			return nil, err
		}

		m = format.MetadataImage(image)
	}

	// If necessary, crop to fit exact size.
	if o.Crop && (o.Width < m.Width || o.Height < m.Height) {
		// Center horizontally
		x := (m.Width - o.Width + 1) / 2
		// Assume faces are higher up vertically
		y := (m.Height - o.Height + 1) / 4

		if err := image.ExtractArea(m.Orientation.Crop(o.Width, o.Height, x, y, m.Width, m.Height)); err != nil {
			return nil, err
		}

		m = format.MetadataImage(image)
	}

	if err = postProcess(image, m.Orientation, shrank, o); err != nil {
		return nil, err
	}

	return format.Save(image, saveOptions)
}

func load(blob []byte, f format.Format, shrink int) (*vips.Image, error) {
	if f == format.Jpeg && shrink > 1 {
		return vips.JpegloadBufferShrink(blob, shrink)
	}

	return f.LoadBytes(blob)
}

func preProcess(image *vips.Image) error {
	_ = image.IccImport()

	space := image.ImageGuessInterpretation()
	if space != vips.InterpretationSRGB && space != vips.InterpretationBW {
		if err := image.Colourspace(vips.InterpretationSRGB); err != nil {
			return err
		}
	}

	if image.HasAlpha() {
		// TODO: Check if image has alpha channel set to 100% opaque
		// and Flatten() it instead.

		// Interpolation of RGB values with an alpha channel isn't
		// safe unless the values are pre-multiplied.  Undo this
		// later.
		if err := image.Premultiply(); err != nil {
			return err
		}
	}

	return nil
}

func postProcess(image *vips.Image, orientation format.Orientation, shrank bool, options Options) error {
	if options.BlurSigma > 0.0 {
		if err := image.Gaussblur(options.BlurSigma); err != nil {
			return err
		}
	}

	if options.Sharpen && shrank {
		if err := image.MildSharpen(); err != nil {
			return err
		}
	}

	if image.HasAlpha() {
		// Assume we pre-multiplied above and undo it after all
		// operations that touch adjacent pixels.
		if err := image.Unpremultiply(); err != nil {
			return err
		}
	}

	// Make sure we generate images with 8 bits per channel. Do this
	// before the rotate to reduce the amount of data that needs to be
	// copied.
	if image.ImageGetBandFormat() != vips.BandFormatUchar {
		if err := image.Cast(vips.BandFormatUchar); err != nil {
			return err
		}
	}

	// Before rotating this will also apply all operations above into a
	// copy of the image.
	return orientation.Apply(image)
}

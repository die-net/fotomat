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

	m = format.MetadataImage(image)

	// A box filter will quickly get us within 2x of the final size.
	shrink = math.Floor(scaleFactor(m.Width, m.Height, iw, ih))
	if shrink >= 2 {
		out, err := image.Shrink(shrink, shrink)
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
		m = format.MetadataImage(image)
	}

	// Do a high-quality resize to scale to final size.
	if iw < m.Width || ih < m.Height {
		factor := scaleFactor(iw, ih, m.Width, m.Height)
		out, err := image.Resize(factor, factor)
		if err != nil {
			return nil, err
		}

		image.Close()
		image = out
		m = format.MetadataImage(image)
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
		m = format.MetadataImage(image)
	}

	image, err = postProcess(image, m.Orientation, shrank, o)
	if err != nil {
		return nil, err
	}

	thumb, err := format.Save(image, saveOptions)
	image.Close()
	return thumb, err
}

func load(blob []byte, f format.Format, shrink int) (*vips.Image, error) {
	if f == format.Jpeg && shrink > 1 {
		return vips.JpegloadBufferShrink(blob, jpegShrink(shrink))
	}

	return f.LoadBytes(blob)
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
		// TODO: Check if image has alpha channel set to 100% opaque
		// and Flatten() it instead.

		// Interpolation of RGB values with an alpha channel isn't
		// safe unless the values are pre-multiplied.  Undo this
		// later.
		out, err := image.Premultiply()
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
	}

	return image, nil
}

func postProcess(image *vips.Image, orientation format.Orientation, shrank bool, options Options) (*vips.Image, error) {
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
		// Assume we pre-multiplied above and undo it after all
		// operations that touch adjacent pixels.
		out, err := image.Unpremultiply()
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
	}

	// Make sure we generate images with 8 bits per channel. Do this
	// before the rotate to reduce the amount of data that needs to be
	// copied.
	if image.ImageGetBandFormat() != vips.BandFormatUchar {
		out, err := image.Cast(vips.BandFormatUchar)
		if err != nil {
			return nil, err
		}
		image.Close()
		image = out
	}

	// Before rotating this will also apply all operations above into a
	// copy of the image.
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

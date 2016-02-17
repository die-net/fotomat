package imager

type Options struct {
	Width                   int
	Height                  int
	Crop                    bool
	Format                  Format
	Quality                 int
	Compression             int
	Lossless                bool
	MaxBufferPixels         int
	LosslessMaxBitsPerPixel int
	Sharpen                 bool
	BlurSigma               float64
	AutoContrast            bool
}

func (o *Options) Check(format Format, width, height int) error {
	// If output format is not set, pick one.
	if o.Format == UnknownFormat {
		switch format {
		case Gif:
			o.Format = Png
		default:
			o.Format = format
		}
	}
	// Is this now a format that can save? If not, error.
	if !o.Format.CanSave() {
		return ErrUnknownFormat
	}

	// If output width or height are not set, use original.
	if o.Width == 0 {
		o.Width = width
	}
	if o.Height == 0 {
		o.Height = height
	}
	// Security: Verify requested width and height.
	if o.Width < 1 || o.Height < 1 {
		return ErrTooSmall
	}
	if o.Width > maxDimension || o.Height > maxDimension {
		return ErrTooBig
	}

	// If set, limit allocated pixels to MaxBufferPixels.  Assume JPEG
	// decoder can pre-scale to 1/8 original width and height.
	scale := 1
	if format == Jpeg {
		scale = 8
	}
	if o.MaxBufferPixels > 0 && width*height > o.MaxBufferPixels*scale*scale {
		return ErrTooBig
	}

	if o.Quality == 0 {
		o.Quality = DefaultQuality
	}
	if o.Quality < 1 || o.Quality > 100 {
		return ErrBadOption
	}

	if o.Compression == 0 {
		o.Compression = DefaultCompression
	}
	if o.Compression < 1 || o.Compression > 9 {
		return ErrBadOption
	}

	if o.BlurSigma < 0.0 || o.BlurSigma > 8.0 {
		return ErrBadOption
	}

	return nil
}

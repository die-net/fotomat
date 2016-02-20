package imager

type Options struct {
	Width           int
	Height          int
	Crop            bool
	MaxBufferPixels int
	Sharpen         bool
	BlurSigma       float64
	AutoContrast    bool
	SaveOptions
}

func (o *Options) Check(m Metadata) error {
	// Security: Limit formats we pass to VIPS to JPEG, PNG, GIF, WEBP.
	if m.Format == UnknownFormat {
		return ErrUnknownFormat
	}

	// Security: Confirm that image sizes are sane.
	if m.Width < minDimension || m.Height < minDimension {
		return ErrTooSmall
	}
	if m.Width > maxDimension || m.Height > maxDimension {
		return ErrTooBig
	}

	// If output format is not set, pick one.
	if o.Format == UnknownFormat {
		switch m.Format {
		case Gif:
			o.Format = Png
		default:
			o.Format = m.Format
		}
	}
	// Is this now a format that can save? If not, error.
	if !o.Format.CanSave() {
		return ErrUnknownFormat
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
		return ErrTooSmall
	}
	if o.Width > maxDimension || o.Height > maxDimension {
		return ErrTooBig
	}
	// If requested crop width or height are larger than original, scale
	// request down to fit within original dimensions.
	if o.Crop && (o.Width > m.Width || o.Height > m.Height) {
		o.Width, o.Height = scaleAspect(o.Width, o.Height, m.Width, m.Height, true)
	}

	// If set, limit allocated pixels to MaxBufferPixels.  Assume JPEG
	// decoder can pre-scale to 1/8 original width and height.
	scale := 1
	if m.Format == Jpeg {
		scale = 8
	}
	if o.MaxBufferPixels > 0 && m.Width*m.Height > o.MaxBufferPixels*scale*scale {
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

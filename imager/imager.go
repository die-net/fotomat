package imager

import (
	"errors"
)

var (
	UnknownFormat = errors.New("Unknown image format")
	TooBig        = errors.New("Image is too wide or tall")
)

type Imager struct {
	blob         []byte
	Width        uint
	Height       uint
	InputFormat  string
	OutputFormat string
	Quality      uint
}

func New(blob []byte) (*Imager, error) {
	// Security: Guess at formats.  Limit formats we pass to ImageMagick to just JPEG, PNG, GIF.
	inputFormat, outputFormat := detectFormats(blob)
	if inputFormat == "" {
		return nil, UnknownFormat
	}

	// Ask ImageMagick to parse metadata.
	width, height, format, err := imageMetaData(blob)
	if err != nil {
		return nil, UnknownFormat
	}

	// Security: Confirm that detectFormat() and imageMagick agreed on format and that
	// image sizes are not likely to wrap shorts (limited to 2<<14-2 intentionally).
	if format != inputFormat || width < 1 || height < 1 {
		return nil, UnknownFormat
	} else if width > 16382 || height > 16382 {
		return nil, TooBig
	}

	img := &Imager{
		blob:         blob,
		Width:        width,
		Height:       height,
		InputFormat:  inputFormat,
		OutputFormat: outputFormat,
		Quality:      85,
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

func (img *Imager) Close() {
	*img = Imager{}
}

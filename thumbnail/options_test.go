package thumbnail

import (
	"github.com/die-net/fotomat/format"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOptionsMetadata(t *testing.T) {
	_, err := Options{}.Check(format.Metadata{})
	assert.Equal(t, err, format.ErrUnknownFormat)

	_, err = Options{}.Check(format.Metadata{Width: 1, Height: 1, Format: format.Jpeg})
	assert.Equal(t, err, ErrTooSmall)

	_, err = Options{}.Check(format.Metadata{Width: 34000, Height: 34000, Format: format.Jpeg})
	assert.Equal(t, err, ErrTooBig)

	_, err = Options{}.Check(format.Metadata{Width: 2, Height: 2, Format: format.Jpeg})
	assert.Equal(t, err, nil)
}

func TestOptionsValidation(t *testing.T) {
	m := format.Metadata{Width: 640, Height: 480, Format: format.Jpeg}

	_, err := Options{BlurSigma: 10}.Check(m)
	assert.Equal(t, err, ErrBadOption)

	_, err = Options{BlurSigma: -1}.Check(m)
	assert.Equal(t, err, ErrBadOption)

	_, err = Options{Width: -1}.Check(m)
	assert.Equal(t, err, ErrTooSmall)

	_, err = Options{Height: -1}.Check(m)
	assert.Equal(t, err, ErrTooSmall)

	_, err = Options{Width: 32767}.Check(m)
	assert.Equal(t, err, ErrTooBig)

	_, err = Options{Height: 32767}.Check(m)
	assert.Equal(t, err, ErrTooBig)
}

func TestOptionsCrop(t *testing.T) {
	m := format.Metadata{Width: 640, Height: 480, Format: format.Jpeg}

	// When cropping, width and height are adjusted to the aspect ratio.
	r, err := Options{Width: 400, Height: 800, Crop: true}.Check(m)
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Width, 240)
	assert.Equal(t, r.Height, 480)

	// Without the Crop flag, we don't adjust dimensions.
	r, err = Options{Width: 400, Height: 800, Crop: false}.Check(m)
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Width, 400)
	assert.Equal(t, r.Height, 800)
}

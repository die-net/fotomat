package imager

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

/*
func TestOptionsDefaults(t *testing.T) {
	r, err := Options{}.Check(format.Metadata{Width: 2, Height: 2, Format: format.Jpeg})
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Quality, 85)
	assert.Equal(t, r.Compression, 6)
}
*/

/*
func TestOptionsFormats(t *testing.T) {
	// Make sure Jpeg is written back as Jpeg by default.
	r, err := Options{}.Check(format.Metadata{Width: 2, Height: 2, Format: format.Jpeg})
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Format, format.Jpeg)

	// Make sure Gif is written as Png by default.
	r, err = Options{}.Check(format.Metadata{Width: 2, Height: 2, Format: format.Gif})
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Format, Png)

	// Disallow explicitly saving as Gif.
	_, err = Options{Format: format.Gif}.Check(format.Metadata{Width: 2, Height: 2, Format: format.Gif})
	assert.Equal(t, err, ErrUnknownFormat)
}
*/

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

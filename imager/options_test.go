// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOptionsMetadata(t *testing.T) {
	_, err := Options{}.Check(Metadata{})
	assert.Equal(t, err, ErrUnknownFormat)

	_, err = Options{}.Check(Metadata{Width: 1, Height: 1, Format: Jpeg})
	assert.Equal(t, err, ErrTooSmall)

	_, err = Options{}.Check(Metadata{Width: 34000, Height: 34000, Format: Jpeg})
	assert.Equal(t, err, ErrTooBig)

	_, err = Options{}.Check(Metadata{Width: 2, Height: 2, Format: Jpeg})
	assert.Equal(t, err, nil)
}

func TestOptionsDefaults(t *testing.T) {
	r, err := Options{}.Check(Metadata{Width: 2, Height: 2, Format: Jpeg})
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Quality, 85)
	assert.Equal(t, r.Compression, 6)
}

func TestOptionsFormats(t *testing.T) {
	// Make sure Jpeg is written back as Jpeg by default.
	r, err := Options{}.Check(Metadata{Width: 2, Height: 2, Format: Jpeg})
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Format, Jpeg)

	// Make sure Gif is written as Png by default.
	r, err = Options{}.Check(Metadata{Width: 2, Height: 2, Format: Gif})
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Format, Png)

	// Disallow explicitly saving as Gif.
	_, err = Options{Format: Gif}.Check(Metadata{Width: 2, Height: 2, Format: Gif})
	assert.Equal(t, err, ErrUnknownFormat)
}

func TestOptionsCrop(t *testing.T) {
	m := Metadata{Width: 640, Height: 480, Format: Jpeg}

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

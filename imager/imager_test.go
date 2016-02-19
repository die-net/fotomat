// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestImageValidation(t *testing.T) {
	// Return ErrUnknownFormat on a text file.
	assert.Equal(t, tryNew("notimage.txt"), ErrUnknownFormat)

	// Return ErrUnknownFormat on a truncated image.
	assert.Equal(t, tryNew("bad.jpg"), ErrUnknownFormat)

	// Refuse to load a 1x1 pixel image.
	assert.Equal(t, tryNew("1px.png"), ErrTooSmall)

	// Load a 2x2 pixel image.
	assert.Nil(t, tryNew("2px.png"))

	// Return ErrTooBig on a 34000x16 PNG image.
	assert.Equal(t, tryNew("34000px.png"), ErrTooBig)

	// Refuse to load a 213328 pixel JPEG image into 1000 pixel buffer.
	// TODO: Add back MaxBufferPixels.
	_, err := Thumbnail(image("watermelon.jpg"), Options{Width: 200, Height: 300, MaxBufferPixels: 1000})
	assert.Equal(t, err, ErrTooBig)

	// Succeed in loading a 213328 pixel JPEG image into 10000 pixel buffer.
	_, err = Thumbnail(image("watermelon.jpg"), Options{Width: 200, Height: 300, MaxBufferPixels: 10000})
	assert.Nil(t, err)
}

func tryNew(filename string) error {
	_, err := Thumbnail(image(filename), Options{Width: 200, Height: 200})
	return err
}

func TestImageThumbnail(t *testing.T) {
	img := image("watermelon.jpg")

	m, err := MetadataBytes(img)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, m.Width, 398)
	assert.Equal(t, m.Height, 536)

	// Verify scaling down to fit completely into box.
	thumb, err := Thumbnail(img, Options{Width: 200, Height: 300})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, Jpeg, 200, 269))
	}

	// Verify scaling down to have width fit.
	thumb, err = Thumbnail(img, Options{Width: 200})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, Jpeg, 200, 269))
	}

	// Verify scaling down to have height fit.
	thumb, err = Thumbnail(img, Options{Height: 300})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, Jpeg, 223, 300))
	}

	// Verify that we don't scale up.
	thumb, err = Thumbnail(img, Options{Width: 2048, Height: 2048})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, Jpeg, 398, 536))
	}
}

func TestImageCrop(t *testing.T) {
	img := image("watermelon.jpg")

	m, err := MetadataBytes(img)
	assert.Nil(t, err)
	assert.Equal(t, m.Width, 398)
	assert.Equal(t, m.Height, 536)

	// Verify cropping to fit.
	thumb, err := Thumbnail(img, Options{Width: 300, Height: 400, Crop: true})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, Jpeg, 300, 400))
	}

	// Verify cropping to fit, too big.
	thumb, err = Thumbnail(img, Options{Width: 2000, Height: 1500, Crop: true})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, Jpeg, 398, 299))
	}
}

func TestImageRotation(t *testing.T) {
	for i := 1; i <= 8; i++ {
		// Verify that New() correctly translates dimensions.
		img := image("orient" + strconv.Itoa(i) + ".jpg")

		m, err := MetadataBytes(img)
		if assert.Nil(t, err) {
			assert.Equal(t, m.Width, 48)
			assert.Equal(t, m.Height, 80)

			// Verify that img.Thumbnail() maintains orientation.
			thumb, err := Thumbnail(img, Options{Width: 40, Height: 40})
			assert.Nil(t, err)
			assert.Nil(t, isSize(thumb, Jpeg, 24, 40))
		}

		// TODO: Figure out how to test crop.
	}
}

func TestImageFormat(t *testing.T) {
	img := image("2px.gif")

	m, err := MetadataBytes(img)
	if assert.Nil(t, err) {
		assert.Equal(t, m.Width, 2)
		assert.Equal(t, m.Height, 3)

		// Verify that we rewrite it as a PNG of the same size.
		thumb, err := Thumbnail(img, Options{Width: 1024, Height: 1024})
		assert.Nil(t, err)
		assert.Nil(t, isSize(thumb, Png, 2, 3))
	}

	img = image("flowers.png")

	m, err = MetadataBytes(img)
	if assert.Nil(t, err) {
		assert.Equal(t, m.Width, 256)
		assert.Equal(t, m.Height, 169)

		// Verify that we rewrite it as JPEG of the same size.
		thumb, err := Thumbnail(img, Options{Width: 1024, Height: 1024, SaveOptions: SaveOptions{LosslessMaxBitsPerPixel: 4}})
		assert.Nil(t, err)
		assert.Nil(t, isSize(thumb, Jpeg, 256, 169))
	}
}

func isSize(image []byte, format Format, width, height int) error {
	m, err := MetadataBytes(image)
	if err != nil {
		return err
	}
	if m.Width != width || m.Height != height {
		return fmt.Errorf("Width %d!=%d or height %d!=%d", m.Width, width, m.Height, height)
	}
	if m.Format != format {
		return fmt.Errorf("Format %s!=%s", m.Format, format)
	}
	return nil
}

func image(filename string) []byte {
	bytes, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		panic(err)
	}

	return bytes
}

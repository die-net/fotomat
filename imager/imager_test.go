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
	assert.Equal(t, tryNew("watermelon.jpg"), ErrTooBig)

	// Succeed in loading a 213328 pixel JPEG image into 10000 pixel buffer.
        // TODO: Add back MaxBufferPixels.
	assert.Nil(t, tryNew("watermelon.jpg"))

	// Load the image when given a larger limit.
        // TODO: Add back MaxBufferPixels.
	assert.Nil(t, tryNew("watermelon.jpg"))
}

func tryNew(filename string) error {
	img, err := New(image(filename))
	if img != nil {
		img.Close()
	}
	return err
}

func TestImageThumbnail(t *testing.T) {
	img, err := New(image("watermelon.jpg"))
	defer img.Close()
	assert.Nil(t, err)
	assert.Equal(t, img.width, 398)
	assert.Equal(t, img.height, 536)

	// Verify scaling down to fit completely into box.
	thumb, err := img.Thumbnail(Options{Width: 200, Height: 300})
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, Jpeg, 200, 269))

	// Verify scaling down to have one side fit into box.
	thumb, err = img.Thumbnail(Options{Width: 200, Height: 300})
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, Jpeg, 223, 300))

	// Verify that we don't scale up.
	thumb, err = img.Thumbnail(Options{Width: 2048, Height: 2048})
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, Jpeg, 398, 536))
}

func TestImageCrop(t *testing.T) {
	img, err := New(image("watermelon.jpg"))
	defer img.Close()
	assert.Nil(t, err)
	assert.Equal(t, img.width, uint(398))
	assert.Equal(t, img.height, uint(536))

	// Verify cropping to fit.
	thumb, err := img.Crop(Options{Width: 300, Height: 400, Crop: true})
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, Jpeg, 300, 400))

	// Verify cropping to fit, too big.
	thumb, err = img.Crop(Options{Width: 2000, Height: 1500, Crop: true})
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, Jpeg, 398, 299))
}

func TestImageRotation(t *testing.T) {
	for i := 1; i <= 8; i++ {
		// Verify that New() correctly translates dimensions.
		img, err := New(image("orient" + strconv.Itoa(i) + ".jpg"))
		defer img.Close()
		assert.Nil(t, err)
		assert.Equal(t, img.width, uint(48))
		assert.Equal(t, img.height, uint(80))

		// Verify that img.Thumbnail() maintains orientation.
		thumb, err := img.Thumbnail(Options{Width: 40, Height: 40})
		assert.Nil(t, err)
		assert.Nil(t, isSize(thumb, Jpeg, 24, 40))

		// TODO: Figure out how to test crop.
	}
}

func TestImageFormat(t *testing.T) {
	img, err := New(image("2px.gif"))
	assert.Nil(t, err)
	assert.Equal(t, img.width, uint(2))
	assert.Equal(t, img.height, uint(3))

	// Verify that we rewrite it as a PNG of the same size.
	thumb, err := img.Thumbnail(Options{Width: 1024, Height: 1024})
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, Png, 2, 3))
	img.Close()

	img, err = New(image("flowers.png"))
	assert.Nil(t, err)
	assert.Equal(t, img.width, uint(256))
	assert.Equal(t, img.height, uint(169))

	// Verify that we rewrite it as JPEG of the same size.
	thumb, err = img.Thumbnail(Options{Width: 1024, Height: 1024})
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, Jpeg, 256, 169))
	img.Close()
}

func isSize(image []byte, format Format, width, height int) error {
	img, err := New(image)
	if err != nil {
		return err
	}
	defer img.Close()
	if width != img.width || height != img.height {
		return fmt.Errorf("Width %d!=%d or height %d!=%d", width, img.width, height, img.height)
	}
	if format != img.format {
		return fmt.Errorf("Format %s!=%s", format, img.format)
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

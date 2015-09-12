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
	// Return UnknownFormat on a text file.
	assert.Equal(t, tryNew("notimage.txt", 1000000), UnknownFormat)

	// Return UnknownFormat on a truncated image.
	assert.Equal(t, tryNew("bad.jpg", 1000000), UnknownFormat)

	// Refuse to load a 1x1 pixel image.
	assert.Equal(t, tryNew("1px.png", 1000000), UnknownFormat)

	// Load a 2x2 pixel image.
	assert.Nil(t, tryNew("2px.png", 1000000))

	// Return TooBig on a 34000x16 image.
	assert.Equal(t, tryNew("34000px.png", 10000000), TooBig)

	// Refuse to load a 213328 pixel image.
	assert.Equal(t, tryNew("watermelon.jpg", 10000), TooBig)

	// Load the image when given a larger limit.
	assert.Nil(t, tryNew("watermelon.jpg", 100000))
}

func tryNew(filename string, maxBufferPixels uint) error {
	img, err := New(image(filename), maxBufferPixels)
        if img != nil {
            img.Close()
        }
	return err
}

func TestImageThumbnail(t *testing.T) {
	img, err := New(image("watermelon.jpg"), 10000000)
	defer img.Close()
	assert.Nil(t, err)
	assert.Equal(t, img.Width, uint(398))
	assert.Equal(t, img.Height, uint(536))

	// Verify scaling down to fit completely into box.
	thumb, err := img.Thumbnail(200, 300, true)
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, "JPEG", 200, 269))

	// Verify scaling down to have one side fit into box.
	thumb, err = img.Thumbnail(200, 300, false)
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, "JPEG", 223, 300))

	// Verify that we don't scale up.
	thumb, err = img.Thumbnail(2048, 2048, true)
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, "JPEG", 398, 536))
}

func TestImageCrop(t *testing.T) {
	img, err := New(image("watermelon.jpg"), 10000000)
	defer img.Close()
	assert.Nil(t, err)
	assert.Equal(t, img.Width, uint(398))
	assert.Equal(t, img.Height, uint(536))

	// Verify cropping to fit.
	thumb, err := img.Crop(300, 400)
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, "JPEG", 300, 400))

	// Verify cropping to fit, too big.
	thumb, err = img.Crop(2000, 1500)
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, "JPEG", 398, 299))
}

func TestImageRotation(t *testing.T) {
        for i := 1; i <= 8; i++ {
                // Verify that New() correctly translates dimensions.
		img, err := New(image("orient"+strconv.Itoa(i)+".jpg"), 10000000)
		defer img.Close()
		assert.Nil(t, err)
		assert.Equal(t, img.Width, uint(48))
		assert.Equal(t, img.Height, uint(80))

                // Verify that img.Thumbnail() maintains orientation.
		thumb, err := img.Thumbnail(40, 40, true)
		assert.Nil(t, err)
		assert.Nil(t, isSize(thumb, "JPEG", 24, 40))

                // TODO: Figure out how to test crop.
        }
}

func TestImageFormat(t *testing.T) {
	img, err := New(image("flowers.png"), 10000000)
	defer img.Close()
	assert.Nil(t, err)
	assert.Equal(t, img.Width, uint(256))
	assert.Equal(t, img.Height, uint(169))

	// Verify that we rewrite it as JPEG of the same size.
	thumb, err := img.Thumbnail(1024, 1024, true)
	assert.Nil(t, err)
	assert.Nil(t, isSize(thumb, "JPEG", 256, 169))
}


func isSize(image []byte, format string, width, height uint) error {
	img, err := New(image, 10000000)
	if err != nil {
		return err
	}
	defer img.Close()
	if width != img.Width || height != img.Height {
		return fmt.Errorf("Width %d!=%d or Height %d!=%d", width, img.Width, height, img.Height)
	}
	if format != img.InputFormat {
		return fmt.Errorf("Format %s!=%s", format, img.InputFormat)
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

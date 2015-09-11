// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
        "fmt"
        "github.com/stretchr/testify/assert"
        "io/ioutil"
        "testing"
)

func TestImageValidation(t *testing.T) {
    // Return UnknownFormat on a text file.
    assert.Equal(t, tryNew("notimage.txt", 1000000), UnknownFormat)

    // Refuse to load a 1x1 pixel image.
    assert.Equal(t, tryNew("1px.png", 1000000), UnknownFormat)

    // Load a 2x2 pixel image.
    assert.Nil(t, tryNew("2px.png", 1000000))

    // Refuse to load a 213328 pixel image.
    assert.Equal(t, tryNew("watermelon.jpg", 10000), TooBig)

    // Load the image when given a larger limit.
    assert.Nil(t, tryNew("watermelon.jpg", 100000))
}

func tryNew(filename string, maxBufferPixels uint) error {
    _, err := New(image(filename), maxBufferPixels)
    return err
}


func TestImageThumbnail(t *testing.T) {
    img, err := New(image("watermelon.jpg"), 10000000)
    assert.Nil(t, err)
    assert.Equal(t, img.Width, uint(398))
    assert.Equal(t, img.Height, uint(536))

    // Verify scaling down to fit completely into box.
    thumb, err := img.Thumbnail(200, 300, true)
    assert.Nil(t, err)
    assert.Nil(t, isSize(thumb, 200, 269))

    // Verify scaling down to have one side fit into box.
    thumb, err = img.Thumbnail(200, 300, false)
    assert.Nil(t, err)
    assert.Nil(t, isSize(thumb, 223, 300))

    // Verify that we don't scale up.
    thumb, err = img.Thumbnail(2048, 2048, true)
    assert.Nil(t, err)
    assert.Nil(t, isSize(thumb, 398, 536))
}

func isSize(image []byte, width, height uint) error {
    img, err := New(image, 10000000)
    if err != nil {
        return err
    }
    if width == img.Width && height == img.Height {
        return nil
    }
    return fmt.Errorf("Width %d!=%d or Height %d!=%d", width, img.Width, height, img.Height)
}

func image(filename string) []byte {
    bytes, err := ioutil.ReadFile("testdata/"+filename)
    if err != nil {
        panic(err)
    }

    return bytes
}

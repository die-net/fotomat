package main

import (
	"flag"
	"fmt"
	"github.com/die-net/fotomat/format"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"testing"
)

var localhost string

func init() {
	// Initialize flags with default values, enable local serving.
	flag.Parse()
	*localImageDirectory = "."
	poolInit(1)
	runtime.GOMAXPROCS(2)

	// Listen on an ephemeral localhost port.
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// Record that address.
	localhost = listen.Addr().String()

	go http.Serve(listen, nil)
}

func TestSuccess(t *testing.T) {
	// Load a 2x3 pixel image.
	assert.Nil(t, isSize("2px.png=s2048x2048", format.Png, 2, 3))

	// Crop JPEG to 200x100.
	assert.Nil(t, isSize("watermelon.jpg=c200x100", format.Jpeg, 200, 100))

	// Scale preview JPEG.
	assert.Nil(t, isSize("watermelon.jpg=ps100x100", format.Jpeg, 75, 100))

	// Crop 3000x2000 PNG to a small preview JPEG.
	assert.Nil(t, isSize("3000px.png=pc16x16", format.Jpeg, 16, 16))
}

func TestResponseErrors(t *testing.T) {
	// Return StatusNotFound on a textfile that doesn't exist.
	assert.Equal(t, status("notfound.txt=s16x16"), http.StatusNotFound)

	// Return StatusUnsupportedMediaType on a text file.
	assert.Equal(t, status("notimage.txt=s16x16"), http.StatusUnsupportedMediaType)

	// Return StatusUnsupportedMediaType on a truncated image.
	assert.Equal(t, status("bad.jpg=s16x16"), http.StatusUnsupportedMediaType)

	// Return StatusUnsupportedMediaType on a 1x1 pixel image.
	assert.Equal(t, status("1px.png=s16x16"), http.StatusUnsupportedMediaType)

	// Return StatusRequestEntityTooLarge on a 34000px image.
	assert.Equal(t, status("34000px.png=s16x16"), http.StatusRequestEntityTooLarge)
}

func TestParameterValidation(t *testing.T) {
	// Test missing parameters.
	assert.Equal(t, status("watermelon.jpg"), http.StatusBadRequest)

	// Test bad operation.
	assert.Equal(t, status("watermelon.jpg=z16x16"), http.StatusBadRequest)

	// Require preview flag to be a prefix.
	assert.Equal(t, status("watermelon.jpg=sp16x16"), http.StatusBadRequest)

	// Test that both scale and crop refuse a 0px width or height.
	assert.Equal(t, status("watermelon.jpg=s0x10"), http.StatusBadRequest)
	assert.Equal(t, status("watermelon.jpg=s10x0"), http.StatusBadRequest)
	assert.Equal(t, status("watermelon.jpg=c0x10"), http.StatusBadRequest)
	assert.Equal(t, status("watermelon.jpg=c10x0"), http.StatusBadRequest)

	// Test that both scale and crop refuse a 2049px width or height.
	assert.Equal(t, status("watermelon.jpg=s2049x16"), http.StatusBadRequest)
	assert.Equal(t, status("watermelon.jpg=s16x2049"), http.StatusBadRequest)
	assert.Equal(t, status("watermelon.jpg=c2049x16"), http.StatusBadRequest)
	assert.Equal(t, status("watermelon.jpg=c16x2049"), http.StatusBadRequest)

	// Refuse repeated scale parameters.
	assert.Equal(t, status("watermelon.jpg=s16x16=s16x16"), http.StatusBadRequest)
}

func isSize(filename string, f format.Format, width, height int) error {
	image, code := fetch(filename)
	if code != 200 {
		return fmt.Errorf("HTTP error %d", code)
	}

	m, err := format.MetadataBytes(image)
	if err != nil {
		return err
	}
	if m.Width != width || m.Height != height {
		return fmt.Errorf("Width %d!=%d or Height %d!=%d", m.Width, width, m.Height, height)
	}
	if m.Format != f {
		return fmt.Errorf("Format %s!=%s", m.Format, f)
	}
	return nil
}

func status(filename string) int {
	_, code := fetch(filename)
	return code
}

func fetch(filename string) ([]byte, int) {
	resp, err := http.Get("http://" + localhost + "/imager/testdata/" + filename)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body, resp.StatusCode
}

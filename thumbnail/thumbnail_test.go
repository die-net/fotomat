package thumbnail

import (
	"fmt"
	"github.com/die-net/fotomat/format"
	"github.com/die-net/fotomat/vips"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	vips.Initialize()
	vips.LeakSet(true)
	r := m.Run()
	vips.ThreadShutdown()
	vips.Shutdown()
	os.Exit(r)
}

func TestValidation(t *testing.T) {
	// Return ErrUnknownFormat on a text file.
	assert.Equal(t, tryNew("notimage.txt"), format.ErrUnknownFormat)

	// Return ErrUnknownFormat on a truncated image.
	assert.Equal(t, tryNew("bad.jpg"), format.ErrUnknownFormat)

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

func TestThumbnail(t *testing.T) {
	img := image("watermelon.jpg")

	m, err := format.MetadataBytes(img)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, m.Width, 398)
	assert.Equal(t, m.Height, 536)

	// Verify scaling down to fit completely into box.
	thumb, err := Thumbnail(img, Options{Width: 200, Height: 300})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 200, 270, false))
	}

	// Verify scaling down to have width fit.
	thumb, err = Thumbnail(img, Options{Width: 200})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 200, 270, false))
	}

	// Verify scaling down to have height fit.
	thumb, err = Thumbnail(img, Options{Height: 300})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 223, 300, false))
	}

	// Verify that we don't scale up.
	thumb, err = Thumbnail(img, Options{Width: 2048, Height: 2048})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 398, 536, false))
	}
}

func TestCrop(t *testing.T) {
	img := image("watermelon.jpg")

	m, err := format.MetadataBytes(img)
	assert.Nil(t, err)
	assert.Equal(t, m.Width, 398)
	assert.Equal(t, m.Height, 536)

	// Verify cropping to fit.
	thumb, err := Thumbnail(img, Options{Width: 300, Height: 400, Crop: true})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 300, 400, false))
	}

	// Verify cropping to fit, too big.
	thumb, err = Thumbnail(img, Options{Width: 2000, Height: 1500, Crop: true})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 398, 299, false))
	}
}

func TestAlpha(t *testing.T) {
	img := image("noalpha.png")
	assert.Nil(t, isSize(img, format.Png, 100, 50, true))

	// Test that we remove the alpha channel from an image that's not using it.
	thumb, err := Thumbnail(img, Options{Width: 100, Height: 100, Save: format.SaveOptions{Format: format.Png}})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Png, 100, 50, false))
	}

	img = image("somealpha.png")
	assert.Nil(t, isSize(img, format.Png, 100, 50, true))

	// Test that we leave the alpha channel for an image that is using it.
	thumb, err = Thumbnail(img, Options{Width: 100, Height: 100, Save: format.SaveOptions{Format: format.Png}})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Png, 100, 50, true))
	}
}

func TestRotation(t *testing.T) {
	for i := 1; i <= 8; i++ {
		// Verify that New() correctly translates dimensions.
		img := image("orient" + strconv.Itoa(i) + ".jpg")

		m, err := format.MetadataBytes(img)
		if !assert.Nil(t, err) {
			continue
		}
		assert.Equal(t, m.Width, 48)
		assert.Equal(t, m.Height, 80)

		// Verify that img.Thumbnail() maintains orientation.
		thumb, err := Thumbnail(img, Options{Width: 40, Height: 40})
		if assert.Nil(t, err) {
			assert.Nil(t, isSize(thumb, format.Jpeg, 24, 40, false))
		}

		// TODO: Figure out how to test crop.
	}
}

func TestConversion(t *testing.T) {
	var formatTest = []struct {
		filename    string
		in          format.Format
		outLossless format.Format
	}{
		{"2px.gif", format.Gif, format.Png},
		{"2px.png", format.Png, format.Png},
		{"2px.jpg", format.Jpeg, format.Jpeg},
		{"2px.webp", format.Webp, format.Png},
	}
	for _, f := range formatTest {
		img := image(f.filename)

		m, err := format.MetadataBytes(img)
		if assert.Nil(t, err, "format: %s", f.in) {
			assert.Equal(t, m.Format, f.in, "format: %s", f.in)
			assert.Equal(t, m.Width, 2, "format: %s", f.in)
			assert.Equal(t, m.Height, 3, "format: %s", f.in)

			for _, of := range []format.Format{format.Png, format.Jpeg, format.Webp} {
				// If we ask for a specific format, it should return that.
				thumb, err := Thumbnail(img, Options{Width: 1024, Height: 1024, Save: format.SaveOptions{Format: of}})
				if assert.Nil(t, err, "formats: %s -> %s", f.in, of) {
					assert.Nil(t, isSize(thumb, of, 2, 3, false), "formats: %s -> %s", f.in, of)
				}
			}
		}

	}
}

func TestScalingJpeg(t *testing.T) {
	testScalingFormat(t, format.Jpeg)
}

func TestScalingPng(t *testing.T) {
	testScalingFormat(t, format.Png)
}

func TestScalingWebp(t *testing.T) {
	testScalingFormat(t, format.Webp)
}

func testScalingFormat(t *testing.T, f format.Format) {
	blob, err := flowersFormat(f)
	if !assert.Nil(t, err) {
		return
	}

	// Try scaling to some difficult sizes and make sure we get the expected size back.
	// We have different code paths for different image formats, so we try for each.
	for _, size := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256} {
		thumb, err := Thumbnail(blob, Options{Width: size, Height: size, Save: format.SaveOptions{Format: f}})
		if assert.Nil(t, err) {
			h := (169*size + 255) / 256
			if h < 1 {
				h = 1
			}
			assert.Nil(t, isSize(thumb, f, size, h, false))
		}
	}
}

func TestBadOptions(t *testing.T) {
	for _, f := range []format.Format{format.Png, format.Jpeg, format.Webp} {
		blob, err := flowersFormat(f)
		if !assert.Nil(t, err, "format: %s", f) {
			continue
		}

		// Try feeding some bad options and make sure we get nothing back.
		for _, of := range []format.Format{format.Png, format.Jpeg, format.Webp} {
			thumb, err := Thumbnail(blob, Options{Width: -1, Height: -1, Save: format.SaveOptions{Format: of}})
			assert.Error(t, err, "format: %s -> %s, size: %d", f, of)
			assert.False(t, thumb != nil, "format: %s -> %s, size: %d", f, of)
		}
	}
}

func BenchmarkThumbnailJpeg_16(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 16, Height: 16})
}

func BenchmarkThumbnailJpeg_32(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 32, Height: 32})
}

func BenchmarkThumbnailJpeg_64(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 64, Height: 64})
}

func BenchmarkThumbnailJpeg_128(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 128, Height: 128})
}

func BenchmarkThumbnailJpeg_256(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 256, Height: 256})
}

func BenchmarkThumbnailJpeg_24(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 24, Height: 24})
}

func BenchmarkThumbnailJpeg_48(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 48, Height: 48})
}

func BenchmarkThumbnailJpeg_96(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 96, Height: 96})
}

func BenchmarkThumbnailJpeg_192(b *testing.B) {
	benchThumbnail(b, format.Jpeg, Options{Width: 192, Height: 192})
}

func BenchmarkThumbnailPng_16(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 16, Height: 16})
}

func BenchmarkThumbnailPng_32(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 32, Height: 32})
}

func BenchmarkThumbnailPng_64(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 64, Height: 64})
}

func BenchmarkThumbnailPng_128(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 128, Height: 128})
}

func BenchmarkThumbnailPng_256(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 256, Height: 256})
}

func BenchmarkThumbnailPng_24(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 24, Height: 24})
}

func BenchmarkThumbnailPng_48(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 48, Height: 48})
}

func BenchmarkThumbnailPng_96(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 96, Height: 96})
}

func BenchmarkThumbnailPng_192(b *testing.B) {
	benchThumbnail(b, format.Png, Options{Width: 192, Height: 192})
}

func BenchmarkThumbnailWebp_16(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 16, Height: 16})
}

func BenchmarkThumbnailWebp_32(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 32, Height: 32})
}

func BenchmarkThumbnailWebp_64(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 64, Height: 64})
}

func BenchmarkThumbnailWebp_128(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 128, Height: 128})
}

func BenchmarkThumbnailWebp_256(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 256, Height: 256})
}

func BenchmarkThumbnailWebp_24(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 24, Height: 24})
}

func BenchmarkThumbnailWebp_48(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 48, Height: 48})
}

func BenchmarkThumbnailWebp_96(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 96, Height: 96})
}

func BenchmarkThumbnailWebp_192(b *testing.B) {
	benchThumbnail(b, format.Webp, Options{Width: 192, Height: 192})
}

func benchThumbnail(b *testing.B, f format.Format, o Options) {
	o.Save.Format = f
	blob, err := flowersFormat(f)
	assert.Nil(b, err)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Thumbnail(blob, o)
			assert.Nil(b, err)
		}
	})
}

func flowersFormat(f format.Format) ([]byte, error) {
	return Thumbnail(image("flowers.png"), Options{Save: format.SaveOptions{Format: f}})
}

func isSize(image []byte, f format.Format, width, height int, alpha bool) error {
	m, err := format.MetadataBytes(image)
	if err != nil {
		return err
	}
	if m.Width != width || m.Height != height {
		return fmt.Errorf("Got %dx%d != want %dx%d", m.Width, m.Height, width, height)
	}
	if m.Format != f {
		return fmt.Errorf("Format %s!=%s", m.Format, f)
	}
	if m.HasAlpha != alpha {
		return fmt.Errorf("HasAlpha %s!=%s", m.HasAlpha, alpha)
	}
	return nil
}

func image(filename string) []byte {
	bytes, err := ioutil.ReadFile("../testdata/" + filename)
	if err != nil {
		panic(err)
	}

	return bytes
}

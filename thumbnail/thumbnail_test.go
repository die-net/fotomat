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

func TestImageValidation(t *testing.T) {
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
	_, err := Thumbnail(image("watermelon.jpg"), Options{Width: 200, Height: 300, MaxBufferPixels: 1000}, format.SaveOptions{})
	assert.Equal(t, err, ErrTooBig)

	// Succeed in loading a 213328 pixel JPEG image into 10000 pixel buffer.
	_, err = Thumbnail(image("watermelon.jpg"), Options{Width: 200, Height: 300, MaxBufferPixels: 10000}, format.SaveOptions{})
	assert.Nil(t, err)
}

func tryNew(filename string) error {
	_, err := Thumbnail(image(filename), Options{Width: 200, Height: 200}, format.SaveOptions{})
	return err
}

func TestImageThumbnail(t *testing.T) {
	img := image("watermelon.jpg")

	m, err := format.MetadataBytes(img)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, m.Width, 398)
	assert.Equal(t, m.Height, 536)

	// Verify scaling down to fit completely into box.
	thumb, err := Thumbnail(img, Options{Width: 200, Height: 300}, format.SaveOptions{})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 200, 270))
	}

	// Verify scaling down to have width fit.
	thumb, err = Thumbnail(img, Options{Width: 200}, format.SaveOptions{})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 200, 270))
	}

	// Verify scaling down to have height fit.
	thumb, err = Thumbnail(img, Options{Height: 300}, format.SaveOptions{})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 223, 300))
	}

	// Verify that we don't scale up.
	thumb, err = Thumbnail(img, Options{Width: 2048, Height: 2048}, format.SaveOptions{})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 398, 536))
	}
}

func TestImageCrop(t *testing.T) {
	img := image("watermelon.jpg")

	m, err := format.MetadataBytes(img)
	assert.Nil(t, err)
	assert.Equal(t, m.Width, 398)
	assert.Equal(t, m.Height, 536)

	// Verify cropping to fit.
	thumb, err := Thumbnail(img, Options{Width: 300, Height: 400, Crop: true}, format.SaveOptions{})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 300, 400))
	}

	// Verify cropping to fit, too big.
	thumb, err = Thumbnail(img, Options{Width: 2000, Height: 1500, Crop: true}, format.SaveOptions{})
	if assert.Nil(t, err) {
		assert.Nil(t, isSize(thumb, format.Jpeg, 398, 299))
	}
}

func TestImageRotation(t *testing.T) {
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
		thumb, err := Thumbnail(img, Options{Width: 40, Height: 40}, format.SaveOptions{})
		if assert.Nil(t, err) {
			assert.Nil(t, isSize(thumb, format.Jpeg, 24, 40))
		}

		// TODO: Figure out how to test crop.
	}
}

func TestImageConversion(t *testing.T) {
	var formatTest = []struct {
		filename    string
		in          format.Format
		outLossless format.Format
		outLossy    format.Format
	}{
		{"2px.gif", format.Gif, format.Png, format.Jpeg},
		{"2px.png", format.Png, format.Png, format.Jpeg},
		{"2px.jpg", format.Jpeg, format.Jpeg, format.Jpeg},
		{"2px.webp", format.Webp, format.Png, format.Jpeg},
	}
	for _, f := range formatTest {
		img := image(f.filename)

		m, err := format.MetadataBytes(img)
		if assert.Nil(t, err, "format: %s", f.in) {
			assert.Equal(t, m.Format, f.in, "format: %s", f.in)
			assert.Equal(t, m.Width, 2, "format: %s", f.in)
			assert.Equal(t, m.Height, 3, "format: %s", f.in)

			// With lossless disabled, verify that we rewrite in the lossy format.
			thumb, err := Thumbnail(img, Options{Width: 1024, Height: 1024}, format.SaveOptions{})
			if assert.Nil(t, err, "format: %s", f.in) {
				assert.Nil(t, isSize(thumb, f.outLossy, 2, 3), "format: %s", f.in)
			}

			// With lossless enabled, verify that we rewrite in the lossless format.
			thumb, err = Thumbnail(img, Options{Width: 1024, Height: 1024}, format.SaveOptions{LosslessMaxBitsPerPixel: 4})
			if assert.Nil(t, err) {
				assert.Nil(t, isSize(thumb, f.outLossless, 2, 3), "format: %s", f.in)
			}
		}

		for _, of := range []format.Format{format.Png, format.Jpeg, format.Webp} {
			// If we ask for a specific format, it should return that.
			thumb, err := Thumbnail(img, Options{Width: 1024, Height: 1024}, format.SaveOptions{Format: of})
			if assert.Nil(t, err, "formats: %s -> %s", f.in, of) {
				assert.Nil(t, isSize(thumb, of, 2, 3), "formats: %s -> %s", f.in, of)
			}
		}
	}
}

func TestImageScalingJpeg(t *testing.T) {
	testImageScalingFormat(t, format.Jpeg)
}

func TestImageScalingPng(t *testing.T) {
	testImageScalingFormat(t, format.Png)
}

func TestImageScalingWebp(t *testing.T) {
	testImageScalingFormat(t, format.Webp)
}

func testImageScalingFormat(t *testing.T, f format.Format) {
	blob, err := flowersFormat(f)
	if !assert.Nil(t, err) {
		return
	}

	// Try scaling to some difficult sizes and make sure we get the expected size back.
	// We have different code paths for different image formats, so we try for each.
	for _, size := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256} {
		thumb, err := Thumbnail(blob, Options{Width: size, Height: size}, format.SaveOptions{Format: f})
		if assert.Nil(t, err) {
			h := (169*size + 255) / 256
			if h < 1 {
				h = 1
			}
			assert.Nil(t, isSize(thumb, f, size, h))
		}
	}
}

func TestImageBadOptions(t *testing.T) {
	for _, f := range []format.Format{format.Png, format.Jpeg, format.Webp} {
		blob, err := flowersFormat(f)
		if !assert.Nil(t, err, "format: %s", f) {
			continue
		}

		// Try feeding some bad options and make sure we get nothing back.
		for _, of := range []format.Format{format.Png, format.Jpeg, format.Webp} {
			thumb, err := Thumbnail(blob, Options{Width: -1, Height: -1}, format.SaveOptions{Format: of})
			assert.Error(t, err, "format: %s -> %s, size: %d", f, of)
			assert.False(t, thumb != nil, "format: %s -> %s, size: %d", f, of)
		}
	}
}

func TestImageSwitchToLossy(t *testing.T) {
	img := image("flowers.png")

	m, err := format.MetadataBytes(img)
	if assert.Nil(t, err) {
		assert.Equal(t, m.Width, 256)
		assert.Equal(t, m.Height, 169)

		// With lossless disabled, we should always return a JPEG.
		thumb, err := Thumbnail(img, Options{Width: 1024, Height: 1024}, format.SaveOptions{})
		assert.Nil(t, err)
		assert.Nil(t, isSize(thumb, format.Jpeg, 256, 169))

		// With lossless set to a high value, we should return a PNG.
		thumb, err = Thumbnail(img, Options{Width: 1024, Height: 1024}, format.SaveOptions{LosslessMaxBitsPerPixel: 20})
		assert.Nil(t, err)
		assert.Nil(t, isSize(thumb, format.Png, 256, 169))

		// Otherwise, make sure that LosslessMaxBitsPerPixel works as expected.
		thumb, err = Thumbnail(img, Options{Width: 1024, Height: 1024}, format.SaveOptions{LosslessMaxBitsPerPixel: 4})
		assert.Nil(t, err)
		assert.Nil(t, isSize(thumb, format.Jpeg, 256, 169))
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
	s := format.SaveOptions{Format: f}
	blob, err := flowersFormat(f)
	assert.Nil(b, err)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Thumbnail(blob, o, s)
			assert.Nil(b, err)
		}
	})
}

func flowersFormat(f format.Format) ([]byte, error) {
	return Thumbnail(image("flowers.png"), Options{}, format.SaveOptions{Format: f})
}

func isSize(image []byte, f format.Format, width, height int) error {
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
	return nil
}

func image(filename string) []byte {
	bytes, err := ioutil.ReadFile("../testdata/" + filename)
	if err != nil {
		panic(err)
	}

	return bytes
}

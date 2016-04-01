package format

import (
	"fmt"
	"github.com/die-net/fotomat/vips"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

const (
	TestdataPath = "../testdata/"
)

func TestMain(m *testing.M) {
	vips.Initialize()
	vips.LeakSet(true)
	r := m.Run()
	vips.ThreadShutdown()
	vips.Shutdown()
	os.Exit(r)
}

func TestMetadataValidation(t *testing.T) {
	// Return ErrUnknownFormat on a text file.
	assert.Equal(t, metadataError("notimage.txt"), ErrUnknownFormat)

	// Return ErrUnknownFormat on a truncated image.
	assert.Equal(t, metadataError("bad.jpg"), ErrUnknownFormat)

	// Load a 2x3 pixel image of each type.
	assert.Nil(t, isSize(image("2px.jpg"), Jpeg, 2, 3))
	assert.Nil(t, isSize(image("2px.png"), Png, 2, 3))
	assert.Nil(t, isSize(image("2px.gif"), Gif, 2, 3))
	assert.Nil(t, isSize(image("2px.webp"), Webp, 2, 3))
}

func metadataError(filename string) error {
	_, err := MetadataBytes(image(filename))
	return err
}

func image(filename string) []byte {
	bytes, err := ioutil.ReadFile(TestdataPath + filename)
	if err != nil {
		panic(err)
	}

	return bytes
}

func isSize(blob []byte, f Format, width, height int) error {
	m, err := MetadataBytes(blob)
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

func TestFormatCanLoad(t *testing.T) {
	assert.Equal(t, "image/jpeg", Jpeg.String())
	assert.True(t, Jpeg.CanLoadBytes())
	assert.True(t, Jpeg.CanLoadFile())

	assert.Equal(t, "application/octet-stream", Unknown.String())

	assert.False(t, Unknown.CanLoadBytes())
	_, err := Unknown.LoadBytes([]byte("foo"))
	assert.Equal(t, ErrInvalidOperation, err)

	assert.False(t, Unknown.CanLoadFile())
	_, err = Unknown.LoadFile("foo")
	assert.Equal(t, ErrInvalidOperation, err)
}

func TestFormatOrientation(t *testing.T) {
	for i := 1; i <= 8; i++ {
		filename := "orient" + strconv.Itoa(i) + ".jpg"

		m, err := Jpeg.MetadataFile(TestdataPath + filename)
		if assert.Nil(t, err) {
			assert.Equal(t, m.Width, 48)
			assert.Equal(t, m.Height, 80)
		}

		thumb := convert(image(filename), SaveOptions{})
		assert.Nil(t, isSize(thumb, Jpeg, 48, 80))
	}
}

func TestFormatCrop(t *testing.T) {
	// TopLeft requires no correction
	x, y, ow, oh := TopLeft.Crop(800, 600, 88, 42, 1024, 768)
	assert.Equal(t, []int{88, 42, 800, 600}, []int{x, y, ow, oh})

	// BottomRight is rotate=180: x=1024-800-88, y=768-600-42
	x, y, ow, oh = BottomRight.Crop(800, 600, 88, 42, 1024, 768)
	assert.Equal(t, []int{136, 126, 800, 600}, []int{x, y, ow, oh})

	// LeftBottom is rotate=270: swap ow, oh. x=768-600-42
	x, y, ow, oh = LeftBottom.Crop(800, 600, 88, 42, 1024, 768)
	assert.Equal(t, []int{126, 88, 600, 800}, []int{x, y, ow, oh})
}

func TestSwitchToLossy(t *testing.T) {
	img := image("flowers.png")

	m, err := MetadataBytes(img)
	if assert.Nil(t, err) {
		assert.Equal(t, m.Width, 256)
		assert.Equal(t, m.Height, 169)

		// With lossless disabled, we should always return a JPEG.
		thumb := convert(img, SaveOptions{})
		assert.Nil(t, isSize(thumb, Jpeg, 256, 169))

		// With lossless enabled, we should return a PNG.
		thumb = convert(img, SaveOptions{Lossless: true})
		assert.Nil(t, isSize(thumb, Png, 256, 169))

		// With lossless and lossIfPhoto enabled, we should return a Jpeg.
		thumb = convert(img, SaveOptions{Lossless: true, LossyIfPhoto: true})
		assert.Nil(t, isSize(thumb, Jpeg, 256, 169))

		// Try saving as lossy webp.
		thumb = convert(img, SaveOptions{AllowWebp: true})
		if assert.Nil(t, isSize(thumb, Webp, 256, 169)) {
			lossyLen := len(thumb)

			// And make sure that lossless webp is larger.
			thumb = convert(img, SaveOptions{AllowWebp: true, Lossless: true})
			assert.Nil(t, isSize(thumb, Webp, 256, 169))
			// TODO: https://github.com/jcupitt/libvips/issues/410
			// assert.NotEqual(t, len(thumb), lossyLen)  // Lossless should be larger

			// Make sure LossyIfPhoto returns lossy.
			thumb = convert(img, SaveOptions{Format: Webp, Lossless: true, LossyIfPhoto: true})
			assert.Nil(t, isSize(thumb, Webp, 256, 169))
			assert.Equal(t, lossyLen, len(thumb))
		}
	}
}

func convert(blob []byte, so SaveOptions) []byte {
	format := DetectFormat(blob)
	img, err := format.LoadBytes(blob)
	if err != nil {
		panic(err)
	}
	defer img.Close()

	DetectOrientation(img).Apply(img)

	blob, err = Save(img, so)
	if err != nil {
		panic(err)
	}
	return blob
}

func BenchmarkMetadataJpeg_2(b *testing.B) {
	benchMetadata(b, "2px.jpg", Jpeg)
}

func BenchmarkMetadataPng_2(b *testing.B) {
	benchMetadata(b, "2px.png", Png)
}

func BenchmarkMetadataWebp_2(b *testing.B) {
	benchMetadata(b, "2px.webp", Webp)
}

func BenchmarkMetadataJpeg_256(b *testing.B) {
	benchMetadata(b, "flowers.png", Jpeg)
}

func BenchmarkMetadataPng_256(b *testing.B) {
	benchMetadata(b, "flowers.png", Png)
}

func BenchmarkMetadataWebp_256(b *testing.B) {
	benchMetadata(b, "flowers.png", Webp)
}

func BenchmarkMetadataJpeg_536(b *testing.B) {
	benchMetadata(b, "watermelon.jpg", Jpeg)
}

func BenchmarkMetadataPng_536(b *testing.B) {
	benchMetadata(b, "watermelon.jpg", Png)
}

func BenchmarkMetadataWebp_536(b *testing.B) {
	benchMetadata(b, "watermelon.jpg", Webp)
}

func BenchmarkMetadataJpeg_3000(b *testing.B) {
	benchMetadata(b, "3000px.png", Jpeg)
}

func BenchmarkMetadataPng_3000(b *testing.B) {
	benchMetadata(b, "3000px.png", Png)
}

func BenchmarkMetadataWebp_3000(b *testing.B) {
	benchMetadata(b, "3000px.png", Webp)
}

func benchMetadata(b *testing.B, filename string, format Format) {
	blob := convert(image(filename), SaveOptions{Format: format})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := MetadataBytes(blob)
			assert.Nil(b, err)
		}
	})
}

func BenchmarkLoadJpeg_2(b *testing.B) {
	benchLoad(b, "2px.jpg", Jpeg)
}

func BenchmarkLoadPng_2(b *testing.B) {
	benchLoad(b, "2px.png", Png)
}

func BenchmarkLoadWebp_2(b *testing.B) {
	benchLoad(b, "2px.webp", Webp)
}

func BenchmarkLoadJpeg_256(b *testing.B) {
	benchLoad(b, "flowers.png", Jpeg)
}

func BenchmarkLoadPng_256(b *testing.B) {
	benchLoad(b, "flowers.png", Png)
}

func BenchmarkLoadWebp_256(b *testing.B) {
	benchLoad(b, "flowers.png", Webp)
}

func BenchmarkLoadJpeg_536(b *testing.B) {
	benchLoad(b, "watermelon.jpg", Jpeg)
}

func BenchmarkLoadPng_536(b *testing.B) {
	benchLoad(b, "watermelon.jpg", Png)
}

func BenchmarkLoadWebp_536(b *testing.B) {
	benchLoad(b, "watermelon.jpg", Webp)
}

func BenchmarkLoadJpeg_3000(b *testing.B) {
	benchLoad(b, "3000px.png", Jpeg)
}

func BenchmarkLoadPng_3000(b *testing.B) {
	benchLoad(b, "3000px.png", Png)
}

func BenchmarkLoadWebp_3000(b *testing.B) {
	benchLoad(b, "3000px.png", Webp)
}

func benchLoad(b *testing.B, filename string, format Format) {
	blob := convert(image(filename), SaveOptions{Format: format})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			img, err := format.LoadBytes(blob)
			if assert.Nil(b, err) {
				// Images are demand loaded. Actually decode all of the pixels.
				assert.Nil(b, img.Write())

				img.Close()
			}
		}
	})
}

func BenchmarkSaveJpeg_2(b *testing.B) {
	benchSave(b, "2px.jpg", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_2(b *testing.B) {
	benchSave(b, "2px.png", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_2(b *testing.B) {
	benchSave(b, "2px.webp", SaveOptions{Format: Webp})
}

func BenchmarkSaveJpeg_256(b *testing.B) {
	benchSave(b, "flowers.png", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_256(b *testing.B) {
	benchSave(b, "flowers.png", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_256(b *testing.B) {
	benchSave(b, "flowers.png", SaveOptions{Format: Webp})
}

func BenchmarkSaveJpeg_536(b *testing.B) {
	benchSave(b, "watermelon.jpg", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_536(b *testing.B) {
	benchSave(b, "watermelon.jpg", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_536(b *testing.B) {
	benchSave(b, "watermelon.jpg", SaveOptions{Format: Webp})
}

func BenchmarkSaveJpeg_3000(b *testing.B) {
	benchSave(b, "3000px.png", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_3000(b *testing.B) {
	benchSave(b, "3000px.png", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_3000(b *testing.B) {
	benchSave(b, "3000px.png", SaveOptions{Format: Webp})
}

func benchSave(b *testing.B, filename string, so SaveOptions) {
	blob := image(filename)
	format := DetectFormat(blob)
	img, err := format.LoadBytes(blob)
	if !assert.Nil(b, err) || !assert.Nil(b, img.Write()) {
		return
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Save(img, so)
			assert.Nil(b, err)
		}
	})

	img.Close()
}

func BenchmarkUseLossless_2(b *testing.B) {
	benchUseLossless(b, "2px.png", SaveOptions{Format: Png, Lossless: true, LossyIfPhoto: true})
}

func BenchmarkUseLossless_256(b *testing.B) {
	benchUseLossless(b, "flowers.png", SaveOptions{Format: Png, Lossless: true, LossyIfPhoto: true})
}

func BenchmarkUseLossless_536(b *testing.B) {
	benchUseLossless(b, "watermelon.jpg", SaveOptions{Format: Png, Lossless: true, LossyIfPhoto: true})
}

func BenchmarkUseLossless_3000(b *testing.B) {
	benchUseLossless(b, "3000px.png", SaveOptions{Format: Png, Lossless: true, LossyIfPhoto: true})
}

func benchUseLossless(b *testing.B, filename string, so SaveOptions) {
	blob := image(filename)
	format := DetectFormat(blob)
	img, err := format.LoadBytes(blob)
	if !assert.Nil(b, err) || !assert.Nil(b, img.Write()) {
		return
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = useLossless(img, so)
		}
	})

	img.Close()
}

func BenchmarkConvertJpeg_2(b *testing.B) {
	benchConvert(b, "2px.jpg", SaveOptions{Format: Jpeg})
}

func BenchmarkConvertPng_2(b *testing.B) {
	benchConvert(b, "2px.png", SaveOptions{Format: Png})
}

func BenchmarkConvertWebp_2(b *testing.B) {
	benchConvert(b, "2px.webp", SaveOptions{Format: Webp})
}

func BenchmarkConvertJpeg_256(b *testing.B) {
	benchConvert(b, "flowers.png", SaveOptions{Format: Jpeg})
}

func BenchmarkConvertPng_256(b *testing.B) {
	benchConvert(b, "flowers.png", SaveOptions{Format: Png})
}

func BenchmarkConvertWebp_256(b *testing.B) {
	benchConvert(b, "flowers.png", SaveOptions{Format: Webp})
}

func BenchmarkConvertJpeg_536(b *testing.B) {
	benchConvert(b, "watermelon.jpg", SaveOptions{Format: Jpeg})
}

func BenchmarkConvertPng_536(b *testing.B) {
	benchConvert(b, "watermelon.jpg", SaveOptions{Format: Png})
}

func BenchmarkConvertWebp_536(b *testing.B) {
	benchConvert(b, "watermelon.jpg", SaveOptions{Format: Webp})
}

func BenchmarkConvertJpeg_3000(b *testing.B) {
	benchConvert(b, "3000px.png", SaveOptions{Format: Jpeg})
}

func BenchmarkConvertPng_3000(b *testing.B) {
	benchConvert(b, "3000px.png", SaveOptions{Format: Png})
}

func BenchmarkConvertWebp_3000(b *testing.B) {
	benchConvert(b, "3000px.png", SaveOptions{Format: Webp})
}

func benchConvert(b *testing.B, filename string, so SaveOptions) {
	blob := convert(image(filename), so)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = convert(blob, so)
		}
	})
}

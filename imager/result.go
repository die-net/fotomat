package imager

import (
	"fmt"
	"github.com/gographics/imagick/imagick"
)

type Result struct {
	wand   *imagick.MagickWand
	img    *Imager
	Width  uint
	Height uint
}

func (img *Imager) NewResult(width, height uint) (*Result, error) {
	result := &Result{
		img:  img,
		wand: imagick.NewMagickWand(),
	}

	if width > 0 && height > 0 {
		// Ask the jpeg decoder to pre-scale for us, down to something at least
		// as big as this.  This is often a huge performance gain.
		s := fmt.Sprintf("%dx%d", width, height)
		if err := result.wand.SetOption("jpeg:size", s); err != nil {
			result.Close()
			return nil, err
		}
	}

	// Decompress the image into a pixel buffer, possibly pre-scaling first.
	if err := result.wand.ReadImageBlob(img.blob); err != nil {
		result.Close()
		return nil, err
	}

	// These may be smaller than img.Width and img.Height if JPEG decoder pre-scaled image.
	result.Width = result.wand.GetImageWidth()
	result.Height = result.wand.GetImageHeight()

	return result, nil
}

func (result *Result) Resize(width, height uint) error {
        // Only use Lanczos if we are shrinking by more than 2.5%
	maxw := result.Width - result.Width / 40
	maxh := result.Height - result.Height / 40

	if width < maxw && height < maxh {
		return result.wand.ResizeImage(width, height, imagick.FILTER_LANCZOS, 0.8)
	} else {
		return result.wand.ResizeImage(width, height, imagick.FILTER_TRIANGLE, 0.0)
	}
}

///     if err = img.wand.CropImage(w, h, x, y); err != nil {

func (result *Result) Get() ([]byte, error) {
	// Remove extraneous metadata.  Photoshop in particular adds a huge XML blob.
	if err := result.wand.StripImage(); err != nil {
		return nil, err
	}

	// Output image format may differ from input format.
	if err := result.wand.SetImageFormat(result.img.OutputFormat); err != nil {
		return nil, err
	}

	if result.img.OutputFormat == "JPEG" {
		if err := result.wand.SetImageCompressionQuality(result.img.Quality); err != nil {
			return nil, err
		}

		// This creates "Progressive JPEGs", which are smaller.  Don't use for non-JPEG.
		if err := result.wand.SetInterlaceScheme(imagick.INTERLACE_LINE); err != nil {
			return nil, err
		}
	}

	// Run the format-specific compressor, return the byte slice.
	return result.wand.GetImageBlob(), nil
}

func (result *Result) Close() {
	// imagick.MagicWand will otherwise leak unless we Destroy.
	result.wand.Destroy()

	result.wand = nil
	result.img = nil
}

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
	shrank bool
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

	// Don't bother to send 16 or 32 bits per channel.
	if err := result.wand.SetImageDepth(8); err != nil {
		result.Close()
		return nil, err
	}

	// Don't preserve data for fully-transparent pixels.
	if err := result.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_BACKGROUND); err != nil {
		result.Close()
		return nil, err
	}

	// These may be smaller than img.Width and img.Height if JPEG decoder pre-scaled image.
	result.Width = result.wand.GetImageWidth()
	result.Height = result.wand.GetImageHeight()

	if result.Width < img.Width && result.Height < img.Height {
		result.shrank = true
	}

	return result, nil
}

func (result *Result) Resize(width, height uint) error {
	// Only use Lanczos if we are shrinking by more than 2.5%.
	filter := imagick.FILTER_TRIANGLE
	if width < result.Width-result.Width/40 && height < result.Height-result.Height/40 {
		filter = imagick.FILTER_LANCZOS
	}

	if err := result.wand.ResizeImage(width, height, filter, 1); err != nil {
		return err
	}

	// Only change dimensions and/or set shrank flag on success.
	result.Width = width
	result.Height = height
	if filter == imagick.FILTER_LANCZOS {
		result.shrank = true
	}

	return nil
}


func (result *Result) Crop(width, height uint) error {
        if width > result.Width || height > result.Height {
             return TooBig
        }

        // Center horizontally
        x := (int(result.Width) - int(width) + 1) / 2
        // Assume faces are higher up vertically
        y := (int(result.Height) - int(height) + 1) / 4

	if err := result.wand.CropImage(width, height, x, y); err != nil {
		return err
        }

        result.Width = width
        result.Height = height

	return nil
}


func (result *Result) Get() ([]byte, error) {
	// If the image shrunk, apply a light sharpening pass
	if result.shrank && result.img.Sharpen {
		if err := result.wand.UnsharpMaskImage(0, 0.8, 0.6, 0); err != nil {
			return nil, err
		}
	}

	// Remove extraneous metadata.
	if err := stripProfilesAndComments(result.wand); err != nil {
		return nil, err
	}

	// Output image format may differ from input format.
	if err := result.wand.SetImageFormat(result.img.OutputFormat); err != nil {
		return nil, err
	}

	switch result.img.OutputFormat {
	case "JPEG":
		if err := result.wand.SetImageCompressionQuality(result.img.JpegQuality); err != nil {
			return nil, err
		}

		// This creates "Progressive JPEGs", which are smaller.
		// Don't use for non-JPEG.
		if err := result.wand.SetInterlaceScheme(imagick.INTERLACE_LINE); err != nil {
			return nil, err
		}
	case "PNG":
		// PNG quality: 95 = Gzip level=9, adaptive strategy=5
		if err := result.wand.SetImageCompressionQuality(95); err != nil {
			return nil, err
		}
	}

	// Run the format-specific compressor, return the byte slice.
	return result.wand.GetImageBlob(), nil
}

func (result *Result) Close() {
	// imagick.MagicWand will otherwise leak unless we wand.Destroy().
	result.wand.Destroy()

	*result = Result{}
}

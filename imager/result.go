// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"fmt"
	"github.com/gographics/imagick/imagick"
)

type Result struct {
	wand        *imagick.MagickWand
	img         *Imager
	Width       uint
	Height      uint
	Orientation Orientation
	shrank      bool
}

func (img *Imager) NewResult(width, height uint) (*Result, error) {
	result := &Result{
		Orientation: *img.Orientation,
		img:         img,
		wand:        imagick.NewMagickWand(),
	}

	// Swap width and height if orientation will be corrected later.
	width, height = result.Orientation.Dimensions(width, height)

	// If we're scaling down, ask the jpeg decoder to pre-scale for us,
	// down to something at least as big as this.  This is often a huge
	// performance gain.
	if width > 0 && height > 0 && img.Width > width && img.Height > height {
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

	// Make sure that we are using the first frame of an animation.
	result.wand.ResetIterator()

	// Reset virtual canvas and position.
	if err := result.wand.ResetImagePage(""); err != nil {
		result.Close()
		return nil, err
	}

	if result.applyColorProfile() {
		// Make sure ImageMagick is aware that this is now sRGB.
		if err := result.wand.SetColorspace(imagick.COLORSPACE_SRGB); err != nil {
			result.Close()
			return nil, err
		}
	} else if result.wand.GetImageColorspace() != imagick.COLORSPACE_SRGB {
		// Switch to sRGB colorspace, the default for the web.
		if err := result.wand.TransformImageColorspace(imagick.COLORSPACE_SRGB); err != nil {
			result.Close()
			return nil, err
		}
	}

	// These may be smaller than img.Width and img.Height if JPEG decoder pre-scaled image.
	result.Width, result.Height = result.Orientation.Dimensions(result.wand.GetImageWidth(), result.wand.GetImageHeight())

	if result.Width < img.Width && result.Height < img.Height {
		result.shrank = true
	}

	return result, nil
}

func (result *Result) applyColorProfile() bool {
	icc := result.wand.GetImageProfile("icc")
	if icc == "" {
		return false // no color profile
	}

	if icc == sRGB_IEC61966_2_1_black_scaled {
		return true // already applied
	}

	// Apply sRGB IEC 61966 2.1 to this image.
	err := result.wand.ProfileImage("icc", []byte(sRGB_IEC61966_2_1_black_scaled))
	return err == nil // did we successfully apply?
}

func (result *Result) Resize(width, height uint) error {
	// Only use Lanczos if we are shrinking by more than 2.5%.
	filter := imagick.FILTER_TRIANGLE
	shrinking := false
	if width < result.Width-result.Width/40 && height < result.Height-result.Height/40 {
		filter = imagick.FILTER_LANCZOS
		shrinking = true
	}

	ow, oh := result.Orientation.Dimensions(width, height)
	if err := result.wand.ResizeImage(ow, oh, filter, 1); err != nil {
		return err
	}

	// Only change dimensions and/or set shrank flag on success.
	result.Width = width
	result.Height = height
	result.shrank = shrinking

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

	ow, oh, ox, oy := result.Orientation.Crop(width, height, x, y, result.Width, result.Height)
	if err := result.wand.CropImage(ow, oh, ox, oy); err != nil {
		return err
	}

	result.Width = width
	result.Height = height

	return nil
}

func (result *Result) Get() ([]byte, error) {
	// If the image shrunk, apply sharpen or blur as requested
	if result.shrank {
		if result.img.Sharpen {
			if err := result.wand.UnsharpMaskImage(0, 0.8, 0.6, 0.05); err != nil {
				return nil, err
			}
		} else if result.img.BlurSigma > 0 {
			if err := result.wand.GaussianBlurImage(0, result.img.BlurSigma); err != nil {
				return nil, err
			}
		}
	}

	// Only save at 8 bits per channel.
	if err := result.wand.SetImageDepth(8); err != nil {
		return nil, err
	}

	// Fix orientation.
	if err := result.Orientation.Fix(result.wand); err != nil {
		return nil, err
	}

	// Stretch contrast if AutoContrast flag set.
	if result.img.AutoContrast {
		if err := result.wand.NormalizeImage(); err != nil {
			return nil, err
		}
	}

	// Remove extraneous metadata and color profiles.
	if err := result.wand.StripImage(); err != nil {
		result.Close()
		return nil, err
	}

	hasAlpha := result.wand.GetImageAlphaChannel()
	if hasAlpha {
		// Don't preserve data for fully-transparent pixels.
		if err := result.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_BACKGROUND); err != nil {
			return nil, err
		}
	}

	if result.img.OutputFormat == "PNG" {
		blob, err := result.compress("PNG", 95, imagick.INTERLACE_NO)
		if err != nil || hasAlpha || (len(blob)-256)*8 <= int(result.Width*result.Height*result.img.PngMaxBitsPerPixel) {
			return blob, err
		}
	}

	// Interlace saves 2-3%, but incurs a few hundred bytes of overhead.
	// This isn't usually beneficial on small images.
	if result.Width*result.Height < 200*200 {
		return result.compress("JPEG", result.img.JpegQuality, imagick.INTERLACE_NO)
	}

	return result.compress("JPEG", result.img.JpegQuality, imagick.INTERLACE_LINE)
}

func (result *Result) compress(format string, quality uint, interlace imagick.InterlaceType) ([]byte, error) {
	// Output image format may differ from input format.
	if err := result.wand.SetImageFormat(format); err != nil {
		return nil, err
	}

	if err := result.wand.SetImageCompressionQuality(quality); err != nil {
		return nil, err
	}

	if err := result.wand.SetInterlaceScheme(interlace); err != nil {
		return nil, err
	}

	// Run the format-specific compressor, return the byte slice.
	return result.wand.GetImageBlob(), nil
}

func (result *Result) Close() {
	// imagick.MagicWand will otherwise leak unless we wand.Destroy().
	result.wand.Destroy()

	*result = Result{}
}

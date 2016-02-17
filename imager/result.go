// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"github.com/die-net/fotomat/vips"
)

type Result struct {
	imager      *Imager
	image       *vips.Image
	options     Options
	width       int
	height      int
	orientation Orientation
	shrank      bool
}

func (imager *Imager) NewResult(width, height int, options Options) (*Result, error) {
	// Swap width and height if orientation will be corrected later.
	width, height = imager.Orientation.Dimensions(width, height)

	image, err := imager.shrinkImage(width, height)
	if err != nil {
		return nil, err
	}

	/*
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
	*/

	result := &Result{
		imager:      imager,
		image:       image,
		options:     options,
		orientation: imager.Orientation,
	}

	// These may be smaller than imager.width and imager.height if JPEG decoder pre-scaled image.
	result.width, result.height = result.orientation.Dimensions(result.image.Xsize(), result.image.Ysize())

	if result.width < imager.Width && result.height < imager.Height {
		result.shrank = true
	}

	return result, nil
}

func (imager *Imager) shrinkImage(width, height int) (*vips.Image, error) {
	shrink := imager.Width / width
	ys := imager.Height / height
	if ys < shrink {
		shrink = ys
	}

	// JPEG decode can shrink by a factor of 2, 4, or 8 with a huge
	// performance boost.
	jpegShrink := 1
	if imager.Format == Jpeg {
		switch {
		case shrink >= 8:
			jpegShrink = 8
		case shrink >= 4:
			jpegShrink = 4
		case shrink >= 2:
			jpegShrink = 2
		}
	}

	var err error
	var image *vips.Image
	if jpegShrink > 1 {
		shrink = shrink / jpegShrink
		image, err = vips.JpegloadBufferShrink(imager.blob, jpegShrink)
	} else {
		image, err = imager.image.Copy()
	}
	if err != nil {
		return nil, err
	}

	// Cleanly shrinking by an integer factor can happen with a fast box filter.
	if shrink > 1 {
		out, err := image.Shrink(float64(shrink), float64(shrink))
		image.Close()

		return out, err
	}

	return image, nil
}

/*
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
*/

func (result *Result) Resize(width, height int) error {
	factor := float64(width) / float64(result.width)
	fy := float64(height) / float64(result.height)
	if fy > factor {
		factor = fy
	}

	interpolate := vips.NewInterpolate("bicubic")
	defer interpolate.Close()

	image, err := result.image.Affine(float64(factor), 0, 0, float64(factor), interpolate)
	if err != nil {
		return err
	}

	result.image.Close()
	result.image = image
	result.width = image.Xsize()
	result.height = image.Ysize()
	result.shrank = true

	return nil
}

func (result *Result) Crop(width, height int) error {
	if width > result.width || height > result.height {
		return ErrTooBig
	}

	// Center horizontally
	x := (int(result.width) - int(width) + 1) / 2
	// Assume faces are higher up vertically
	y := (int(result.height) - int(height) + 1) / 4

	ow, oh, ox, oy := result.orientation.Crop(width, height, x, y, result.width, result.height)

	image, err := result.image.ExtractArea(ox, oy, ow, oh)
	if err != nil {
		return err
	}

	result.image.Close()
	result.image = image
	result.width = width
	result.height = height

	return nil
}

func (result *Result) Get() ([]byte, error) {
	// If the image shrunk, apply sharpen or blur as requested
	/*
		if result.shrank {
			if result.imager.Sharpen {
				if err := result.wand.UnsharpMaskImage(0, 0.8, 0.6, 0.05); err != nil {
					return nil, err
				}
			} else if result.imager.BlurSigma > 0 {
				if err := result.wand.GaussianBlurImage(0, result.imager.BlurSigma); err != nil {
					return nil, err
				}
			}
		}
	*/

	// Only save at 8 bits per channel.
	/*
		if err := result.wand.SetImageDepth(8); err != nil {
			return nil, err
		}
	*/

	// Fix orientation.
	image, err := result.orientation.Apply(result.image)
	if err != nil {
		return nil, err
	}
	if image != nil {
		result.image.Close()
		result.image = image
		result.orientation = TopLeft
		result.width = image.Xsize()
		result.height = image.Ysize()
	}

	// Stretch contrast if AutoContrast flag set.
	/*
		if result.imager.AutoContrast {
			if err := result.wand.NormalizeImage(); err != nil {
				return nil, err
			}
		}

	*/
	// Remove extraneous metadata and color profiles.
	/*
		if err := result.wand.StripImage(); err != nil {
			result.Close()
			return nil, err
		}
	*/

	return result.options.Format.Save(result.image, result.options.SaveOptions)

	/*
		hasAlpha := result.wand.GetImageAlphaChannel()
		if hasAlpha {
			// Don't preserve data for fully-transparent pixels.
			if err := result.wand.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_BACKGROUND); err != nil {
				return nil, err
			}
		}
	*/
}

func (result *Result) Close() {
	result.image.Close()

	*result = Result{}
}

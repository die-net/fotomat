// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"github.com/gographics/imagick/imagick"
	"net/http"
)

func detectFormats(blob []byte) (string, string) {
	switch http.DetectContentType(blob) {
	case "image/jpeg":
		return "JPEG", "JPEG"
	case "image/png":
		return "PNG", "PNG"
	case "image/gif":
		return "GIF", "PNG"
	default:
		return "", ""
	}
}

func imageMetaData(blob []byte) (uint, uint, imagick.OrientationType, string, error) {
	// Allocate a temporary wand.
	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	// Get just metadata about the image, don't decode.
	if err := wand.PingImageBlob(blob); err != nil {
		return 0, 0, imagick.ORIENTATION_UNDEFINED, "", err
	}

	width := wand.GetImageWidth()
	height := wand.GetImageHeight()
	orientation := wand.GetImageOrientation()

	if orientationSwapsDimensions(orientation) {
		width, height = height, width
	}

	return width, height, orientation, wand.GetImageFormat(), nil
}

func orientationSwapsDimensions(orientation imagick.OrientationType) bool {
	switch orientation {
	case imagick.ORIENTATION_LEFT_TOP,
		imagick.ORIENTATION_RIGHT_TOP,
		imagick.ORIENTATION_RIGHT_BOTTOM,
		imagick.ORIENTATION_LEFT_BOTTOM:
		return true
	default:
		return false
	}
}

// Scale original (width, height) to result (width, height), maintaining aspect ratio.
// If within=true, fit completely within result, leaving empty space if necessary.
func scaleAspect(ow, oh, rw, rh uint, within bool) (uint, uint) {
	// Scale aspect ratio using integer math, avoiding floating point
	// errors.

	wp := ow * rh
	hp := oh * rw

	if within == (wp < hp) {
		rw = (wp + oh/2) / oh
	} else {
		rh = (hp + ow/2) / ow
	}

	return rw, rh
}

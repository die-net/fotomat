package imager

import (
        "github.com/die-net/fotomat/imager/imagick"
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

func imageMetaData(blob []byte) (uint, uint, string, error) {
	// Allocate a temporary wand.
	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	// Get just metadata about the image, don't decode.
	if err := wand.PingImageBlob(blob); err != nil {
		return 0, 0, "", err
	}

	return wand.GetImageWidth(), wand.GetImageHeight(), wand.GetImageFormat(), nil
}

func stripProfilesAndComments(wand *imagick.MagickWand) error {
	for _, name := range wand.GetImageProfiles("*") {
		// Remove everything except important color profiles.
		if name != "icc" && name != "icm" {
			wand.RemoveImageProfile(name)
		}
	}

	// Remove unnecessary comments.
	return wand.DeleteImageProperty("comment")
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

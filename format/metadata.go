package format

import (
	"github.com/die-net/fotomat/v2/vips"
)

// Metadata is the currently-known metadata about an Image.
type Metadata struct {
	Width       int
	Height      int
	Format      Format
	Orientation Orientation
	HasAlpha    bool
}

// MetadataBytes parses an image byte slice and returns Metadata or an error.
func MetadataBytes(blob []byte) (Metadata, error) {
	format := DetectFormat(blob)
	if format == Unknown {
		return Metadata{}, ErrUnknownFormat
	}

	return format.MetadataBytes(blob)
}

// MetadataBytes parses an image byte slice in known format and returns Metadata or an error.
func (format Format) MetadataBytes(blob []byte) (Metadata, error) {
	image, err := format.LoadBytes(blob)
	if err != nil {
		return Metadata{}, ErrUnknownFormat
	}

	defer image.Close()

	return metadataImageFormat(image, format), nil
}

// MetadataFile parses an image file in known format and returns Metadata or an error.
func (format Format) MetadataFile(filename string) (Metadata, error) {
	image, err := format.LoadFile(filename)
	if err != nil {
		return Metadata{}, err
	}

	defer image.Close()

	return metadataImageFormat(image, format), nil
}

func metadataImageFormat(image *vips.Image, format Format) Metadata {
	m := MetadataImage(image)
	m.Format = format
	return m
}

// MetadataImage returns Metadata from an Image. Format is always unset.
func MetadataImage(image *vips.Image) Metadata {
	o := DetectOrientation(image)
	w, h := o.Dimensions(image.Xsize(), image.Ysize())
	if w <= 0 || h <= 0 {
		panic("Invalid image dimensions.")
	}
	return Metadata{Width: w, Height: h, Orientation: o, HasAlpha: image.HasAlpha()}
}

package format

import (
	"bytes"
	"github.com/die-net/fotomat/vips"
	"image/gif"
)

type Metadata struct {
	Width       int
	Height      int
	Format      Format
	Orientation Orientation
	HasAlpha    bool
}

func MetadataBytes(blob []byte) (Metadata, error) {
	format := DetectFormat(blob)
	if format == Unknown {
		return Metadata{}, ErrUnknownFormat
	}

	return format.MetadataBytes(blob)
}

func (format Format) MetadataBytes(blob []byte) (Metadata, error) {
	if metadata := formatInfo[format].metadata; metadata != nil {
		return metadata(blob)
	}

	image, err := format.LoadBytes(blob)
	if err != nil {
		return Metadata{}, ErrUnknownFormat
	}

	defer image.Close()

	return metadataImageFormat(image, format), nil
}

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

func MetadataImage(image *vips.Image) Metadata {
	o := DetectOrientation(image)
	w, h := o.Dimensions(image.Xsize(), image.Ysize())
	if w <= 0 || h <= 0 {
		panic("Invalid image dimensions.")
	}
	return Metadata{Width: w, Height: h, Orientation: o, HasAlpha: image.HasAlpha()}
}

// vips.MagickloadBuffer completely decodes the image, which is slow and
// unsafe, as we can't check the size before decode. Use Go's GIF reader
// to fetch metadata instead.
func metadataGif(blob []byte) (Metadata, error) {
	c, err := gif.DecodeConfig(bytes.NewReader(blob))
	if err != nil {
		return Metadata{}, err
	}
	return Metadata{Width: c.Width, Height: c.Height, Orientation: Undefined, Format: Gif}, nil
}

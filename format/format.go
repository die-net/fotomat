package format

import (
	"errors"
	"github.com/die-net/fotomat/vips"
	"net/http"
)

var (
	// ErrInvalidOperation is returned for an invalid operation on this image format.
	ErrInvalidOperation = errors.New("Invalid operation")
	// ErrUnknownFormat is returned when the given image is in an unknown format.
	ErrUnknownFormat = errors.New("Unknown image format")
)

// Format of compressed image.
type Format int

// Format of compressed image.
const (
	Unknown Format = iota
	Jpeg
	Png
	Gif
	Webp
)

var formatInfo = []struct {
	mime      string
	loadFile  func(filename string) (*vips.Image, error)
	loadBytes func([]byte) (*vips.Image, error)
}{
	{mime: "application/octet-stream", loadFile: nil, loadBytes: nil},
	{mime: "image/jpeg", loadFile: vips.Jpegload, loadBytes: vips.JpegloadBuffer},
	{mime: "image/png", loadFile: vips.Pngload, loadBytes: vips.PngloadBuffer},
	{mime: "image/gif", loadFile: vips.Gifload, loadBytes: vips.GifloadBuffer},
	{mime: "image/webp", loadFile: vips.Webpload, loadBytes: vips.WebploadBuffer},
}

// DetectFormat detects the Format of the supplied byte slice.
func DetectFormat(blob []byte) Format {
	mime := http.DetectContentType(blob)

	for format, info := range formatInfo {
		if info.mime == mime {
			return Format(format)
		}
	}

	return Unknown
}

// String returns the mime type of given image format.
func (format Format) String() string {
	return formatInfo[format].mime
}

// CanLoadFile returns true if we know how to load this format from a file.
func (format Format) CanLoadFile() bool {
	return formatInfo[format].loadFile != nil
}

// CanLoadBytes returns true if we know how to load this format from a byte slice.
func (format Format) CanLoadBytes() bool {
	return formatInfo[format].loadBytes != nil
}

// LoadFile loads a file in a given Format and returns an Image.
func (format Format) LoadFile(filename string) (*vips.Image, error) {
	loadFile := formatInfo[format].loadFile
	if loadFile == nil {
		return nil, ErrInvalidOperation
	}

	return loadFile(filename)
}

// LoadBytes loads byte slice in a given format and returns an Image.
func (format Format) LoadBytes(blob []byte) (*vips.Image, error) {
	loadBytes := formatInfo[format].loadBytes
	if loadBytes == nil {
		return nil, ErrInvalidOperation
	}

	return loadBytes(blob)
}

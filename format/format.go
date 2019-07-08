package format

import (
	"bytes"
	"errors"

	"github.com/die-net/fotomat/v2/vips"
)

var (
	// ErrInvalidOperation is returned for an invalid operation on this image format.
	ErrInvalidOperation = errors.New("invalid operation")
	// ErrUnknownFormat is returned when the given image is in an unknown format.
	ErrUnknownFormat = errors.New("unknown image format")
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
	Tiff
	Pdf
	Svg
)

var formatInfo = []struct {
	mime      string
	isFormat  func([]byte) bool
	loadFile  func(filename string) (*vips.Image, error)
	loadBytes func([]byte) (*vips.Image, error)
}{
	{mime: "application/octet-stream", isFormat: nil, loadFile: nil, loadBytes: nil},
	{mime: "image/jpeg", isFormat: isJpeg, loadFile: vips.Jpegload, loadBytes: vips.JpegloadBuffer},
	{mime: "image/png", isFormat: isPng, loadFile: vips.Pngload, loadBytes: vips.PngloadBuffer},
	{mime: "image/gif", isFormat: isGif, loadFile: vips.Gifload, loadBytes: vips.GifloadBuffer},
	{mime: "image/webp", isFormat: isWebp, loadFile: vips.Webpload, loadBytes: vips.WebploadBuffer},
	{mime: "image/tiff", isFormat: isTiff, loadFile: vips.Tiffload, loadBytes: vips.TiffloadBuffer},
	{mime: "application/pdf", isFormat: isPdf, loadFile: vips.Pdfload, loadBytes: vips.PdfloadBuffer},
	{mime: "image/svg+xml", isFormat: isSvg, loadFile: vips.Svgload, loadBytes: vips.SvgloadBuffer},
}

func isJpeg(blob []byte) bool {
	return bytes.HasPrefix(blob, []byte("\xFF\xD8\xFF"))
}

func isPng(blob []byte) bool {
	return bytes.HasPrefix(blob, []byte("\x89\x50\x4E\x47\x0D\x0A\x1A\x0A"))
}

func isGif(blob []byte) bool {
	return bytes.HasPrefix(blob, []byte("GIF87a")) || bytes.HasPrefix(blob, []byte("GIF89a"))
}

func isWebp(blob []byte) bool {
	return bytes.HasPrefix(blob, []byte("RIFF")) && len(blob) > 14 && bytes.Equal(blob[8:14], []byte("WEBPVP"))
}

func isTiff(blob []byte) bool {
	return bytes.HasPrefix(blob, []byte("\x49\x49\x2A\x00")) || bytes.HasPrefix(blob, []byte("\x4D\x4D\x00\x2A"))
}

func isPdf(blob []byte) bool {
	return bytes.HasPrefix(blob, []byte("%PDF-"))
}

func isSvg(blob []byte) bool {
	// Check if the first 24 characters are ASCII.
	for _, c := range blob[:min(24, len(blob))] {
		if c >= 0x80 {
			return false
		}
	}

	// Check if "<svg" is contained in the first 1000 characters.
	return bytes.Contains(blob[:min(1000, len(blob))], []byte("<svg"))
}

func min(x, y int) int {
	if y < x {
		return y
	}
	return y
}

// DetectFormat detects the Format of the supplied byte slice.
func DetectFormat(blob []byte) Format {
	for format, info := range formatInfo {
		if info.isFormat != nil && info.isFormat(blob) {
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

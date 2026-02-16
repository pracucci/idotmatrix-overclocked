package graphic

import (
	"bytes"
	"fmt"
	"image/gif"
)

// Display constants
const (
	DisplayWidth  = 64
	DisplayHeight = 64
	BufferSize    = DisplayWidth * DisplayHeight * 3
)

// ImageType represents the type of image (static or animated).
type ImageType int

const (
	ImageTypeStatic   ImageType = iota
	ImageTypeAnimated
)

// Image holds either static RGB data or animated GIF data.
type Image struct {
	Type       ImageType
	StaticData []byte   // Raw RGB (64*64*3 bytes) for static images
	GIFData    *gif.GIF // For animated images
}

// RawBytes returns the raw RGB bytes for static images.
// Returns an error if called on an animated image.
func (img *Image) RawBytes() ([]byte, error) {
	if img.Type != ImageTypeStatic {
		return nil, fmt.Errorf("RawBytes called on non-static image")
	}
	return img.StaticData, nil
}

// GIFBytes encodes the GIF data to bytes.
// Returns an error if called on a static image.
func (img *Image) GIFBytes() ([]byte, error) {
	if img.Type != ImageTypeAnimated {
		return nil, fmt.Errorf("GIFBytes called on non-animated image")
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, img.GIFData); err != nil {
		return nil, fmt.Errorf("failed to encode GIF: %w", err)
	}
	return buf.Bytes(), nil
}

// NewBuffer creates a new 64x64x3 black buffer.
func NewBuffer() []byte {
	return make([]byte, BufferSize)
}

// NewBufferWithColor creates a new 64x64x3 buffer filled with the given color.
func NewBufferWithColor(color Color) []byte {
	buf := make([]byte, BufferSize)
	for i := 0; i < DisplayWidth*DisplayHeight; i++ {
		offset := i * 3
		buf[offset] = color[0]
		buf[offset+1] = color[1]
		buf[offset+2] = color[2]
	}
	return buf
}

// SetPixel sets a single pixel in the RGB image buffer.
// Coordinates outside the display bounds are silently ignored.
func SetPixel(buf []byte, x, y int, color Color) {
	if x < 0 || x >= DisplayWidth || y < 0 || y >= DisplayHeight {
		return
	}
	offset := (y*DisplayWidth + x) * 3
	buf[offset] = color[0]
	buf[offset+1] = color[1]
	buf[offset+2] = color[2]
}

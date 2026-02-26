package graphic

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
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

// ImageToRGB converts an image.Image to a 64x64x3 RGB buffer.
func ImageToRGB(img image.Image) []byte {
	buf := make([]byte, BufferSize)
	bounds := img.Bounds()
	for y := 0; y < DisplayHeight; y++ {
		for x := 0; x < DisplayWidth; x++ {
			srcX := bounds.Min.X + x
			srcY := bounds.Min.Y + y
			if srcX < bounds.Max.X && srcY < bounds.Max.Y {
				r, g, b, _ := img.At(srcX, srcY).RGBA()
				offset := (y*DisplayWidth + x) * 3
				buf[offset] = uint8(r >> 8)
				buf[offset+1] = uint8(g >> 8)
				buf[offset+2] = uint8(b >> 8)
			}
		}
	}
	return buf
}

// MirrorBufferHorizontal mirrors a 64x64 RGB buffer horizontally.
func MirrorBufferHorizontal(buf []byte) []byte {
	mirrored := make([]byte, BufferSize)
	for y := 0; y < DisplayHeight; y++ {
		for x := 0; x < DisplayWidth; x++ {
			srcOffset := (y*DisplayWidth + x) * 3
			dstX := DisplayWidth - 1 - x
			dstOffset := (y*DisplayWidth + dstX) * 3
			mirrored[dstOffset] = buf[srcOffset]
			mirrored[dstOffset+1] = buf[srcOffset+1]
			mirrored[dstOffset+2] = buf[srcOffset+2]
		}
	}
	return mirrored
}

// MirrorGIFHorizontal mirrors all frames of a GIF horizontally.
func MirrorGIFHorizontal(g *gif.GIF) *gif.GIF {
	mirrored := &gif.GIF{
		Image:     make([]*image.Paletted, len(g.Image)),
		Delay:     g.Delay,
		LoopCount: g.LoopCount,
		Disposal:  g.Disposal,
		Config:    g.Config,
	}
	for i, frame := range g.Image {
		bounds := frame.Bounds()
		newFrame := image.NewPaletted(bounds, frame.Palette)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				dstX := bounds.Max.X - 1 - x
				newFrame.SetColorIndex(dstX, y, frame.ColorIndexAt(x, y))
			}
		}
		mirrored.Image[i] = newFrame
	}
	return mirrored
}

// Mirror returns a new Image with horizontal mirroring applied.
func (img *Image) Mirror() *Image {
	if img.Type == ImageTypeStatic {
		return &Image{
			Type:       ImageTypeStatic,
			StaticData: MirrorBufferHorizontal(img.StaticData),
		}
	}
	return &Image{
		Type:    ImageTypeAnimated,
		GIFData: MirrorGIFHorizontal(img.GIFData),
	}
}

// RGBToPaletted converts an RGB buffer to a paletted image for GIF encoding.
func RGBToPaletted(rgbBuf []byte) *image.Paletted {
	rgba := image.NewRGBA(image.Rect(0, 0, DisplayWidth, DisplayHeight))
	for y := 0; y < DisplayHeight; y++ {
		for x := 0; x < DisplayWidth; x++ {
			offset := (y*DisplayWidth + x) * 3
			rgba.Set(x, y, color.RGBA{
				R: rgbBuf[offset],
				G: rgbBuf[offset+1],
				B: rgbBuf[offset+2],
				A: 255,
			})
		}
	}
	paletted := image.NewPaletted(rgba.Bounds(), palette.Plan9)
	draw.Draw(paletted, paletted.Bounds(), rgba, image.Point{}, draw.Src)
	return paletted
}

// AdjustBrightnessBuffer adjusts brightness of a 64x64 RGB buffer.
// brightness is 0-100 where 100 = no change, 0 = black.
// Returns the input unchanged if brightness >= 100.
func AdjustBrightnessBuffer(buf []byte, brightness int) []byte {
	if brightness >= 100 {
		return buf // No change needed
	}

	factor := float64(brightness) / 100.0
	adjusted := make([]byte, BufferSize)
	for i := 0; i < len(buf); i++ {
		adjusted[i] = uint8(float64(buf[i]) * factor)
	}
	return adjusted
}

// AdjustBrightnessGIF adjusts brightness of all frames in a GIF.
// brightness is 0-100 where 100 = no change, 0 = black.
func AdjustBrightnessGIF(g *gif.GIF, brightness int) *gif.GIF {
	if brightness >= 100 {
		return g // No change needed
	}

	factor := float64(brightness) / 100.0
	adjusted := &gif.GIF{
		Image:     make([]*image.Paletted, len(g.Image)),
		Delay:     g.Delay,
		LoopCount: g.LoopCount,
		Disposal:  g.Disposal,
		Config:    g.Config,
	}

	for i, frame := range g.Image {
		bounds := frame.Bounds()
		newPalette := adjustPaletteBrightness(frame.Palette, factor)
		newFrame := image.NewPaletted(bounds, newPalette)
		copy(newFrame.Pix, frame.Pix) // Copy pixel indices directly
		adjusted.Image[i] = newFrame
	}
	return adjusted
}

func adjustPaletteBrightness(p color.Palette, factor float64) color.Palette {
	newPalette := make(color.Palette, len(p))
	for i, c := range p {
		r, g, b, a := c.RGBA()
		newPalette[i] = color.RGBA{
			R: uint8(float64(r>>8) * factor),
			G: uint8(float64(g>>8) * factor),
			B: uint8(float64(b>>8) * factor),
			A: uint8(a >> 8),
		}
	}
	return newPalette
}

// AdjustBrightness returns a new Image with brightness adjusted.
// brightness is 0-100 where 100 = no change, 0 = black.
func (img *Image) AdjustBrightness(brightness int) *Image {
	if brightness >= 100 {
		return img // No change
	}
	if img.Type == ImageTypeStatic {
		return &Image{
			Type:       ImageTypeStatic,
			StaticData: AdjustBrightnessBuffer(img.StaticData, brightness),
		}
	}
	return &Image{
		Type:    ImageTypeAnimated,
		GIFData: AdjustBrightnessGIF(img.GIFData, brightness),
	}
}

package graphic

import (
	"image"
	"image/color"
	"image/gif"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuffer(t *testing.T) {
	buf := NewBuffer()

	// Verify correct size
	assert.Len(t, buf, BufferSize)

	// Verify all bytes are zero (black)
	for i, b := range buf {
		if b != 0 {
			assert.Fail(t, "buffer not zeroed", "NewBuffer()[%d] = %d, want 0", i, b)
			break
		}
	}
}

func TestNewBufferWithColor(t *testing.T) {
	c := Color{100, 150, 200}
	buf := NewBufferWithColor(c)

	// Verify correct size
	assert.Len(t, buf, BufferSize)

	// Verify first pixel has correct RGB values
	assert.Equal(t, c[0], buf[0], "first pixel R")
	assert.Equal(t, c[1], buf[1], "first pixel G")
	assert.Equal(t, c[2], buf[2], "first pixel B")

	// Verify last pixel has correct RGB values
	lastOffset := BufferSize - 3
	assert.Equal(t, c[0], buf[lastOffset], "last pixel R")
	assert.Equal(t, c[1], buf[lastOffset+1], "last pixel G")
	assert.Equal(t, c[2], buf[lastOffset+2], "last pixel B")
}

func TestSetPixel(t *testing.T) {
	c := Color{255, 128, 64}

	t.Run("pixel at (0,0)", func(t *testing.T) {
		buf := NewBuffer()
		SetPixel(buf, 0, 0, c)

		// Offset for (0,0) is 0
		assert.Equal(t, c[0], buf[0])
		assert.Equal(t, c[1], buf[1])
		assert.Equal(t, c[2], buf[2])
	})

	t.Run("pixel at (63,63)", func(t *testing.T) {
		buf := NewBuffer()
		SetPixel(buf, 63, 63, c)

		// Offset for (63,63) is (63*64 + 63) * 3 = 12285
		offset := (63*DisplayWidth + 63) * 3
		assert.Equal(t, c[0], buf[offset])
		assert.Equal(t, c[1], buf[offset+1])
		assert.Equal(t, c[2], buf[offset+2])
	})

	t.Run("out-of-bounds coordinates don't panic", func(t *testing.T) {
		buf := NewBuffer()

		// These should not panic
		SetPixel(buf, -1, 0, c)
		SetPixel(buf, 64, 0, c)
		SetPixel(buf, 0, -1, c)
		SetPixel(buf, 0, 64, c)

		// Buffer should remain all zeros (no writes occurred)
		for i, b := range buf {
			if b != 0 {
				assert.Fail(t, "out-of-bounds write", "Out-of-bounds SetPixel modified buffer at index %d", i)
				break
			}
		}
	})
}

func TestShadowFor(t *testing.T) {
	t.Run("known color returns mapped shadow", func(t *testing.T) {
		shadow := ShadowFor(Red)
		assert.Equal(t, DarkRed, shadow)
	})

	t.Run("unknown color uses fallback", func(t *testing.T) {
		unknown := Color{100, 50, 25}
		shadow := ShadowFor(unknown)

		// Fallback divides each RGB component by 5
		expected := Color{20, 10, 5}
		assert.Equal(t, expected, shadow)
	})
}

func TestImageRawBytes(t *testing.T) {
	t.Run("static image returns StaticData", func(t *testing.T) {
		staticData := NewBufferWithColor(Red)
		img := &Image{
			Type:       ImageTypeStatic,
			StaticData: staticData,
		}

		data, err := img.RawBytes()
		require.NoError(t, err)
		assert.Len(t, data, BufferSize)
	})

	t.Run("animated image returns error", func(t *testing.T) {
		img := &Image{
			Type:    ImageTypeAnimated,
			GIFData: &gif.GIF{},
		}

		_, err := img.RawBytes()
		assert.Error(t, err)
	})
}

func TestImageGIFBytes(t *testing.T) {
	t.Run("animated image returns encoded bytes", func(t *testing.T) {
		// Create a minimal valid GIF
		bounds := image.Rect(0, 0, 2, 2)
		frame := image.NewPaletted(bounds, color.Palette{color.Black, color.White})
		gifData := &gif.GIF{
			Image: []*image.Paletted{frame},
			Delay: []int{10},
		}

		img := &Image{
			Type:    ImageTypeAnimated,
			GIFData: gifData,
		}

		data, err := img.GIFBytes()
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("static image returns error", func(t *testing.T) {
		img := &Image{
			Type:       ImageTypeStatic,
			StaticData: NewBuffer(),
		}

		_, err := img.GIFBytes()
		assert.Error(t, err)
	})
}

func TestMirrorBufferHorizontal(t *testing.T) {
	t.Run("mirrors pixels horizontally", func(t *testing.T) {
		buf := NewBuffer()

		// Set a red pixel at (0, 0) - left edge
		SetPixel(buf, 0, 0, Red)
		// Set a blue pixel at (10, 5)
		SetPixel(buf, 10, 5, Blue)

		mirrored := MirrorBufferHorizontal(buf)

		// Red pixel should now be at (63, 0) - right edge
		offset := (0*DisplayWidth + 63) * 3
		assert.Equal(t, Red[0], mirrored[offset], "mirrored red R")
		assert.Equal(t, Red[1], mirrored[offset+1], "mirrored red G")
		assert.Equal(t, Red[2], mirrored[offset+2], "mirrored red B")

		// Blue pixel at (10, 5) should now be at (53, 5)
		offset = (5*DisplayWidth + 53) * 3
		assert.Equal(t, Blue[0], mirrored[offset], "mirrored blue R")
		assert.Equal(t, Blue[1], mirrored[offset+1], "mirrored blue G")
		assert.Equal(t, Blue[2], mirrored[offset+2], "mirrored blue B")

		// Original position (10, 5) should be black
		offset = (5*DisplayWidth + 10) * 3
		assert.Equal(t, uint8(0), mirrored[offset], "original position should be black")
	})

	t.Run("double mirror returns original", func(t *testing.T) {
		buf := NewBuffer()
		SetPixel(buf, 10, 20, Color{100, 150, 200})
		SetPixel(buf, 30, 40, Color{50, 60, 70})

		mirrored := MirrorBufferHorizontal(buf)
		doubleMirrored := MirrorBufferHorizontal(mirrored)

		assert.Equal(t, buf, doubleMirrored)
	})

	t.Run("returns new buffer, doesn't modify original", func(t *testing.T) {
		buf := NewBuffer()
		SetPixel(buf, 0, 0, Red)
		originalCopy := make([]byte, len(buf))
		copy(originalCopy, buf)

		_ = MirrorBufferHorizontal(buf)

		assert.Equal(t, originalCopy, buf, "original buffer should not be modified")
	})
}

func TestMirrorGIFHorizontal(t *testing.T) {
	t.Run("mirrors single frame GIF", func(t *testing.T) {
		bounds := image.Rect(0, 0, 4, 4)
		palette := color.Palette{color.Black, color.White}
		frame := image.NewPaletted(bounds, palette)

		// Set white pixel at (0, 0)
		frame.SetColorIndex(0, 0, 1)

		g := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 0,
		}

		mirrored := MirrorGIFHorizontal(g)

		// White pixel should now be at (3, 0)
		assert.Equal(t, uint8(1), mirrored.Image[0].ColorIndexAt(3, 0), "white pixel should be mirrored to right edge")
		assert.Equal(t, uint8(0), mirrored.Image[0].ColorIndexAt(0, 0), "original position should be black")
	})

	t.Run("preserves GIF metadata", func(t *testing.T) {
		bounds := image.Rect(0, 0, 2, 2)
		palette := color.Palette{color.Black, color.White}
		frame1 := image.NewPaletted(bounds, palette)
		frame2 := image.NewPaletted(bounds, palette)

		g := &gif.GIF{
			Image:     []*image.Paletted{frame1, frame2},
			Delay:     []int{10, 20},
			LoopCount: 5,
			Disposal:  []byte{gif.DisposalNone, gif.DisposalBackground},
		}

		mirrored := MirrorGIFHorizontal(g)

		assert.Len(t, mirrored.Image, 2, "should have same number of frames")
		assert.Equal(t, g.Delay, mirrored.Delay, "delays should be preserved")
		assert.Equal(t, g.LoopCount, mirrored.LoopCount, "loop count should be preserved")
		assert.Equal(t, g.Disposal, mirrored.Disposal, "disposal should be preserved")
	})

	t.Run("double mirror returns equivalent", func(t *testing.T) {
		bounds := image.Rect(0, 0, 4, 4)
		palette := color.Palette{color.Black, color.White, color.RGBA{255, 0, 0, 255}}
		frame := image.NewPaletted(bounds, palette)
		frame.SetColorIndex(0, 0, 1)
		frame.SetColorIndex(1, 2, 2)

		g := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 0,
		}

		mirrored := MirrorGIFHorizontal(g)
		doubleMirrored := MirrorGIFHorizontal(mirrored)

		// Compare pixel values
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				assert.Equal(t, g.Image[0].ColorIndexAt(x, y), doubleMirrored.Image[0].ColorIndexAt(x, y),
					"pixel at (%d, %d) should match after double mirror", x, y)
			}
		}
	})

	t.Run("returns new GIF, doesn't modify original", func(t *testing.T) {
		bounds := image.Rect(0, 0, 2, 2)
		palette := color.Palette{color.Black, color.White}
		frame := image.NewPaletted(bounds, palette)
		frame.SetColorIndex(0, 0, 1)

		g := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 0,
		}

		// Store original value
		originalValue := g.Image[0].ColorIndexAt(0, 0)

		_ = MirrorGIFHorizontal(g)

		assert.Equal(t, originalValue, g.Image[0].ColorIndexAt(0, 0), "original GIF should not be modified")
	})
}

func TestAdjustBrightnessBuffer(t *testing.T) {
	t.Run("brightness 100 returns input unchanged", func(t *testing.T) {
		buf := NewBufferWithColor(Color{100, 150, 200})
		result := AdjustBrightnessBuffer(buf, 100)
		// Should return same underlying array
		assert.Equal(t, &buf[0], &result[0], "should return same slice when brightness is 100")
	})

	t.Run("brightness 50 halves all values", func(t *testing.T) {
		buf := NewBufferWithColor(Color{100, 150, 200})
		result := AdjustBrightnessBuffer(buf, 50)

		// Check first pixel
		assert.Equal(t, uint8(50), result[0], "R should be halved")
		assert.Equal(t, uint8(75), result[1], "G should be halved")
		assert.Equal(t, uint8(100), result[2], "B should be halved")
	})

	t.Run("brightness 0 produces black", func(t *testing.T) {
		buf := NewBufferWithColor(Color{255, 255, 255})
		result := AdjustBrightnessBuffer(buf, 0)

		for i, b := range result {
			if b != 0 {
				assert.Fail(t, "buffer not zeroed", "AdjustBrightnessBuffer(0)[%d] = %d, want 0", i, b)
				break
			}
		}
	})

	t.Run("brightness above 100 returns input unchanged", func(t *testing.T) {
		buf := NewBufferWithColor(Color{100, 150, 200})
		result := AdjustBrightnessBuffer(buf, 150)
		// Should return same underlying array
		assert.Equal(t, &buf[0], &result[0], "should return same slice when brightness > 100")
	})
}

func TestAdjustBrightnessGIF(t *testing.T) {
	t.Run("brightness 100 returns input unchanged", func(t *testing.T) {
		bounds := image.Rect(0, 0, 2, 2)
		palette := color.Palette{color.Black, color.White}
		frame := image.NewPaletted(bounds, palette)

		g := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 5,
		}

		result := AdjustBrightnessGIF(g, 100)
		assert.Same(t, g, result, "should return same GIF when brightness is 100")
	})

	t.Run("brightness 50 adjusts palette colors", func(t *testing.T) {
		bounds := image.Rect(0, 0, 2, 2)
		palette := color.Palette{
			color.RGBA{R: 100, G: 200, B: 150, A: 255},
			color.RGBA{R: 50, G: 100, B: 75, A: 255},
		}
		frame := image.NewPaletted(bounds, palette)

		g := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 0,
		}

		result := AdjustBrightnessGIF(g, 50)

		// Check adjusted palette - convert RGBA to 8-bit for comparison
		r, gr, b, a := result.Image[0].Palette[0].RGBA()
		assert.Equal(t, uint8(50), uint8(r>>8), "R should be halved")
		assert.Equal(t, uint8(100), uint8(gr>>8), "G should be halved")
		assert.Equal(t, uint8(75), uint8(b>>8), "B should be halved")
		assert.Equal(t, uint8(255), uint8(a>>8), "A should be unchanged")
	})

	t.Run("preserves GIF metadata", func(t *testing.T) {
		bounds := image.Rect(0, 0, 2, 2)
		palette := color.Palette{color.Black, color.White}
		frame1 := image.NewPaletted(bounds, palette)
		frame2 := image.NewPaletted(bounds, palette)

		g := &gif.GIF{
			Image:     []*image.Paletted{frame1, frame2},
			Delay:     []int{10, 20},
			LoopCount: 5,
			Disposal:  []byte{gif.DisposalNone, gif.DisposalBackground},
		}

		result := AdjustBrightnessGIF(g, 50)

		assert.Len(t, result.Image, 2, "should have same number of frames")
		assert.Equal(t, g.Delay, result.Delay, "delays should be preserved")
		assert.Equal(t, g.LoopCount, result.LoopCount, "loop count should be preserved")
		assert.Equal(t, g.Disposal, result.Disposal, "disposal should be preserved")
	})
}

func TestImageAdjustBrightness(t *testing.T) {
	t.Run("static image with brightness 100 returns same image", func(t *testing.T) {
		buf := NewBufferWithColor(Red)
		img := &Image{
			Type:       ImageTypeStatic,
			StaticData: buf,
		}

		result := img.AdjustBrightness(100)
		assert.Same(t, img, result, "should return same image when brightness is 100")
	})

	t.Run("static image brightness adjustment", func(t *testing.T) {
		buf := NewBufferWithColor(Color{100, 150, 200})
		img := &Image{
			Type:       ImageTypeStatic,
			StaticData: buf,
		}

		result := img.AdjustBrightness(50)

		assert.Equal(t, ImageTypeStatic, result.Type)
		assert.Equal(t, uint8(50), result.StaticData[0], "R should be halved")
		assert.Equal(t, uint8(75), result.StaticData[1], "G should be halved")
		assert.Equal(t, uint8(100), result.StaticData[2], "B should be halved")
	})

	t.Run("animated image with brightness 100 returns same image", func(t *testing.T) {
		bounds := image.Rect(0, 0, 2, 2)
		palette := color.Palette{color.Black, color.White}
		frame := image.NewPaletted(bounds, palette)

		gifData := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 0,
		}

		img := &Image{
			Type:    ImageTypeAnimated,
			GIFData: gifData,
		}

		result := img.AdjustBrightness(100)
		assert.Same(t, img, result, "should return same image when brightness is 100")
	})

	t.Run("animated image brightness adjustment", func(t *testing.T) {
		bounds := image.Rect(0, 0, 2, 2)
		palette := color.Palette{color.RGBA{R: 100, G: 200, B: 150, A: 255}}
		frame := image.NewPaletted(bounds, palette)

		gifData := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 0,
		}

		img := &Image{
			Type:    ImageTypeAnimated,
			GIFData: gifData,
		}

		result := img.AdjustBrightness(50)

		assert.Equal(t, ImageTypeAnimated, result.Type)
		r, g, b, _ := result.GIFData.Image[0].Palette[0].RGBA()
		assert.Equal(t, uint8(50), uint8(r>>8), "R should be halved")
		assert.Equal(t, uint8(100), uint8(g>>8), "G should be halved")
		assert.Equal(t, uint8(75), uint8(b>>8), "B should be halved")
	})

	t.Run("doesn't modify original static image", func(t *testing.T) {
		buf := NewBufferWithColor(Color{100, 150, 200})
		img := &Image{
			Type:       ImageTypeStatic,
			StaticData: buf,
		}

		originalFirstPixelR := img.StaticData[0]

		_ = img.AdjustBrightness(50)

		assert.Equal(t, originalFirstPixelR, img.StaticData[0], "original image should not be modified")
	})
}

func TestImageMirror(t *testing.T) {
	t.Run("mirrors static image", func(t *testing.T) {
		buf := NewBuffer()
		SetPixel(buf, 0, 0, Red)

		img := &Image{
			Type:       ImageTypeStatic,
			StaticData: buf,
		}

		mirrored := img.Mirror()

		assert.Equal(t, ImageTypeStatic, mirrored.Type)
		// Red pixel should be at (63, 0)
		offset := (0*DisplayWidth + 63) * 3
		assert.Equal(t, Red[0], mirrored.StaticData[offset])
	})

	t.Run("mirrors animated image", func(t *testing.T) {
		bounds := image.Rect(0, 0, 4, 4)
		palette := color.Palette{color.Black, color.White}
		frame := image.NewPaletted(bounds, palette)
		frame.SetColorIndex(0, 0, 1)

		gifData := &gif.GIF{
			Image:     []*image.Paletted{frame},
			Delay:     []int{10},
			LoopCount: 0,
		}

		img := &Image{
			Type:    ImageTypeAnimated,
			GIFData: gifData,
		}

		mirrored := img.Mirror()

		assert.Equal(t, ImageTypeAnimated, mirrored.Type)
		assert.Equal(t, uint8(1), mirrored.GIFData.Image[0].ColorIndexAt(3, 0))
	})

	t.Run("doesn't modify original", func(t *testing.T) {
		buf := NewBuffer()
		SetPixel(buf, 0, 0, Red)

		img := &Image{
			Type:       ImageTypeStatic,
			StaticData: buf,
		}

		originalFirstPixelR := img.StaticData[0]

		_ = img.Mirror()

		assert.Equal(t, originalFirstPixelR, img.StaticData[0], "original image should not be modified")
	})
}

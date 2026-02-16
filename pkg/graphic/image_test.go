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

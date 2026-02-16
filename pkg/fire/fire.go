package fire

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"math/rand"
	"time"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

const (
	numFrames   = 32
	frameDelay  = 5 // 50ms per frame (delay is in 1/100s)
	paletteSize = 37
)

// DOOM fire palette: black → red → orange → yellow → white
var palette = []color.RGBA{
	{0x07, 0x07, 0x07, 0xFF},
	{0x1F, 0x07, 0x07, 0xFF},
	{0x2F, 0x0F, 0x07, 0xFF},
	{0x47, 0x0F, 0x07, 0xFF},
	{0x57, 0x17, 0x07, 0xFF},
	{0x67, 0x1F, 0x07, 0xFF},
	{0x77, 0x1F, 0x07, 0xFF},
	{0x8F, 0x27, 0x07, 0xFF},
	{0x9F, 0x2F, 0x07, 0xFF},
	{0xAF, 0x3F, 0x07, 0xFF},
	{0xBF, 0x47, 0x07, 0xFF},
	{0xC7, 0x47, 0x07, 0xFF},
	{0xDF, 0x4F, 0x07, 0xFF},
	{0xDF, 0x57, 0x07, 0xFF},
	{0xDF, 0x57, 0x07, 0xFF},
	{0xD7, 0x5F, 0x07, 0xFF},
	{0xD7, 0x5F, 0x07, 0xFF},
	{0xD7, 0x67, 0x0F, 0xFF},
	{0xCF, 0x6F, 0x0F, 0xFF},
	{0xCF, 0x77, 0x0F, 0xFF},
	{0xCF, 0x7F, 0x0F, 0xFF},
	{0xCF, 0x87, 0x17, 0xFF},
	{0xC7, 0x87, 0x17, 0xFF},
	{0xC7, 0x8F, 0x17, 0xFF},
	{0xC7, 0x97, 0x1F, 0xFF},
	{0xBF, 0x9F, 0x1F, 0xFF},
	{0xBF, 0x9F, 0x1F, 0xFF},
	{0xBF, 0xA7, 0x27, 0xFF},
	{0xBF, 0xA7, 0x27, 0xFF},
	{0xBF, 0xAF, 0x2F, 0xFF},
	{0xB7, 0xAF, 0x2F, 0xFF},
	{0xB7, 0xB7, 0x2F, 0xFF},
	{0xB7, 0xB7, 0x37, 0xFF},
	{0xCF, 0xCF, 0x6F, 0xFF},
	{0xDF, 0xDF, 0x9F, 0xFF},
	{0xEF, 0xEF, 0xC7, 0xFF},
	{0xF0, 0xE0, 0x50, 0xFF}, // warm yellow (heat source)
}

// GenerateGIF generates a DOOM-style fire animation GIF.
func GenerateGIF() []byte {
	rand.Seed(time.Now().UnixNano())

	displaySize := graphic.DisplayWidth

	// Initialize fire buffer
	firePixels := make([]int, displaySize*displaySize)

	// Set bottom row to max heat (white)
	for x := 0; x < displaySize; x++ {
		firePixels[(displaySize-1)*displaySize+x] = paletteSize - 1
	}

	// Warmup: run simulation until fire reaches steady state
	// This ensures the GIF starts with flames already burning
	for i := 0; i < 200; i++ {
		spreadFireFrame(firePixels, displaySize)
	}

	// Build color palette for GIF
	gifPalette := make(color.Palette, paletteSize)
	for i, c := range palette {
		gifPalette[i] = c
	}

	frames := make([]*image.Paletted, numFrames)
	delays := make([]int, numFrames)

	for t := 0; t < numFrames; t++ {
		// Simulate fire spread multiple times per frame
		for i := 0; i < 4; i++ {
			spreadFireFrame(firePixels, displaySize)
		}

		// Create frame from buffer
		frame := image.NewPaletted(image.Rect(0, 0, displaySize, displaySize), gifPalette)
		for y := 0; y < displaySize; y++ {
			for x := 0; x < displaySize; x++ {
				idx := firePixels[y*displaySize+x]
				frame.SetColorIndex(x, y, uint8(idx))
			}
		}
		frames[t] = frame
		delays[t] = frameDelay
	}

	g := &gif.GIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: 0,
	}
	var buf bytes.Buffer
	gif.EncodeAll(&buf, g)
	return buf.Bytes()
}

func spreadFireFrame(firePixels []int, displaySize int) {
	for x := 0; x < displaySize; x++ {
		for y := 1; y < displaySize; y++ {
			spreadFire(firePixels, y*displaySize+x, displaySize)
		}
	}
}

func spreadFire(firePixels []int, src int, displaySize int) {
	pixel := firePixels[src]
	if pixel == 0 {
		firePixels[src-displaySize] = 0
		return
	}
	randVal := rand.Intn(8)
	dst := src - randVal + 1
	if dst < displaySize {
		dst = displaySize
	}
	// Decay rate: average 1.0 per row so flames die out in ~36 pixels
	// Values: 0,0,1,1,1,1,2,2 -> average 8/8 = 1.0
	decay := (randVal + 2) / 4
	newVal := pixel - decay
	if newVal < 0 {
		newVal = 0
	}
	firePixels[dst-displaySize] = newVal
}

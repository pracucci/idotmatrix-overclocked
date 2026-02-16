package grot

import (
	"bytes"
	"image"
	"image/gif"
	"image/png"
	"math"
	"math/rand"

	"github.com/pracucci/idotmatrix-overclocked/pkg/assets"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// Matrix animation constants
const (
	matrixFrameCount = 64 // More frames for smoother looping
	matrixFrameDelay = 3  // 30ms per frame (faster to compensate)
	matrixColumns    = 12
	matrixCharWidth  = 4 // 3 pixels + 1 spacing
	matrixCharHeight = 6 // 5 pixels + 1 spacing
	matrixTailLen    = 4 // Number of characters in tail
	matrixRngSeed    = 77

	// Image block constants for dissolution effect
	blockWidth  = 3
	blockHeight = 5
	blockCount  = 30  // Fewer blocks than pixels (they're larger)
	blockSeed   = 123 // Seed for block selection
)

// Tiny 3x5 pixel font for Matrix characters (katakana-inspired + digits)
var matrixChars = [][]string{
	// Simple geometric shapes that evoke katakana
	{"###", "#.#", "###", "#.#", "#.#"}, // 0
	{".#.", "##.", ".#.", ".#.", "###"}, // 1
	{"###", "..#", "###", "#..", "###"}, // 2
	{"###", "..#", "###", "..#", "###"}, // 3
	{"#.#", "#.#", "###", "..#", "..#"}, // 4
	{"###", "#..", "###", "..#", "###"}, // 5
	{"#..", "#..", "###", "#.#", "###"}, // 6
	{"###", "..#", ".#.", ".#.", ".#."}, // 7
	{"###", "#.#", "###", "#.#", "###"}, // 8
	{"###", "#.#", "###", "..#", "..#"}, // 9
	{".#.", "#.#", "###", "#.#", "#.#"}, // A-like
	{"##.", "#.#", "##.", "#.#", "##."}, // B-like
	{"###", "#..", "#..", "#..", "###"}, // C-like
	{".#.", "#.#", "#.#", "#.#", ".#."}, // O-like
	{"#.#", "#.#", ".#.", ".#.", ".#."}, // Y-like
	{"###", ".#.", ".#.", ".#.", ".#."}, // T-like
	{"#.#", "###", "#.#", "#.#", "#.#"}, // H-like
	{"###", "#..", "##.", "#..", "###"}, // E-like
	{".#.", ".#.", ".#.", ".#.", ".#."}, // I-like
	{"#..", "#..", "#..", "#..", "###"}, // L-like
}

// Matrix green colors (bright head to dark tail)
var matrixGreens = []graphic.Color{
	{180, 255, 180}, // White-green (head highlight)
	{0, 255, 0},     // Bright green
	{0, 180, 0},     // Medium green
	{0, 120, 0},     // Dim green
	{0, 60, 0},      // Very dim green
}

// matrixColumn represents a falling column of characters
type matrixColumn struct {
	x         int     // X position (pixel)
	startY    float64 // Starting Y offset (in character units)
	speed     float64 // Fall speed (characters per frame)
	baseChars []int   // Base character indices for deterministic selection
}

// imageBlock represents a 3x5 block that detaches from the base image and falls
type imageBlock struct {
	srcX, srcY int                             // Top-left position in base image
	pixels     [blockHeight][blockWidth]graphic.Color // 3x5 pixel colors
	startFrame int                             // When this block starts falling (0-63)
	speed      float64                         // Fall speed (pixels per frame)
	fallFrames int                             // Pre-calculated frames to exit screen
}

// GenerateMatrix creates a matrix-style animation over the base image.
func GenerateMatrix() (*graphic.Image, error) {
	// Load base image
	baseData, err := assets.Grot.ReadFile("grot/matrix-base.png")
	if err != nil {
		return nil, err
	}

	baseImg, err := png.Decode(bytes.NewReader(baseData))
	if err != nil {
		return nil, err
	}

	// Convert base image to RGB buffer for easy manipulation
	baseRGB := graphic.ImageToRGB(baseImg)

	// Initialize image blocks for dissolution effect
	imageBlocks := initImageBlocks(baseRGB)

	// Initialize random generator with fixed seed for deterministic animation
	rng := rand.New(rand.NewSource(matrixRngSeed))

	numCharRows := graphic.DisplayHeight / matrixCharHeight

	// Calculate cycle length and speed for seamless looping
	// Columns need to travel exactly cycleLength positions in matrixFrameCount frames
	cycleLength := numCharRows + matrixTailLen
	speed := float64(cycleLength) / float64(matrixFrameCount)

	// Initialize columns with staggered starting positions
	columns := make([]*matrixColumn, matrixColumns)
	for i := 0; i < matrixColumns; i++ {
		columns[i] = newMatrixColumn(rng, i, cycleLength, speed)
	}

	var frames []*image.Paletted
	var delays []int

	for frame := 0; frame < matrixFrameCount; frame++ {
		// Start with base image
		buf := make([]byte, len(baseRGB))
		copy(buf, baseRGB)

		// Apply source dimming for blocks (modifies buf in-place for source regions)
		// and draw falling blocks on top
		drawImageBlocks(buf, baseRGB, imageBlocks, frame)

		// Draw matrix columns
		for _, col := range columns {
			drawMatrixColumn(buf, baseRGB, col, frame, numCharRows, cycleLength)
		}

		frames = append(frames, graphic.RGBToPaletted(buf))
		delays = append(delays, matrixFrameDelay)
	}

	return &graphic.Image{
		Type: graphic.ImageTypeAnimated,
		GIFData: &gif.GIF{
			Image:     frames,
			Delay:     delays,
			LoopCount: 0, // Loop forever
		},
	}, nil
}

// newMatrixColumn creates a new falling column with deterministic properties
func newMatrixColumn(rng *rand.Rand, colIndex int, cycleLength int, speed float64) *matrixColumn {
	// Fixed x position (evenly spaced across the screen)
	x := colIndex*5 + 1

	// Randomize starting Y position across the full cycle
	startY := rng.Float64() * float64(cycleLength)

	// Pre-generate base characters for deterministic character selection
	baseChars := make([]int, cycleLength+matrixTailLen+2)
	for i := range baseChars {
		baseChars[i] = rng.Intn(len(matrixChars))
	}

	return &matrixColumn{
		x:         x,
		startY:    startY,
		speed:     speed,
		baseChars: baseChars,
	}
}

// drawMatrixColumn draws a column of characters with fading tail
func drawMatrixColumn(buf, baseRGB []byte, col *matrixColumn, frame int, numCharRows int, cycleLength int) {
	// Calculate head position using cyclic modulo arithmetic for seamless looping
	headY := math.Mod(col.startY+float64(frame)*col.speed, float64(cycleLength))
	headCharY := int(headY)

	// Draw from head through tail
	for i := 0; i <= matrixTailLen; i++ {
		charY := headCharY - i
		if charY < 0 || charY >= numCharRows {
			continue
		}

		// Deterministic character selection based on position and frame
		// Characters change every 4 frames for a subtle flicker effect
		charBaseIdx := (charY + frame/4) % len(col.baseChars)
		charIdx := col.baseChars[charBaseIdx]

		// Choose color based on position (head is brightest)
		var charColor graphic.Color
		if i < len(matrixGreens) {
			charColor = matrixGreens[i]
		} else {
			charColor = matrixGreens[len(matrixGreens)-1]
		}

		// Draw the character
		pixelY := charY * matrixCharHeight
		drawMatrixChar(buf, baseRGB, col.x, pixelY, charIdx, charColor)
	}
}

// drawMatrixChar draws a single 3x5 character at the given position
func drawMatrixChar(buf, baseRGB []byte, x, y, charIdx int, charColor graphic.Color) {
	if charIdx < 0 || charIdx >= len(matrixChars) {
		return
	}

	char := matrixChars[charIdx]
	for row := 0; row < 5; row++ {
		if row >= len(char) {
			continue
		}
		for col := 0; col < 3; col++ {
			if col >= len(char[row]) {
				continue
			}
			if char[row][col] == '#' {
				px := x + col
				py := y + row

				if px < 0 || px >= graphic.DisplayWidth || py < 0 || py >= graphic.DisplayHeight {
					continue
				}

				// Check base image brightness
				offset := (py*graphic.DisplayWidth + px) * 3
				baseR := baseRGB[offset]
				baseG := baseRGB[offset+1]
				baseB := baseRGB[offset+2]
				brightness := int(baseR) + int(baseG) + int(baseB)

				if brightness < 100 {
					// Dark area: show full character
					graphic.SetPixel(buf, px, py, charColor)
				} else {
					// Light area: blend with base
					blendFactor := 0.5
					blendedR := uint8(float64(baseR)*(1-blendFactor) + float64(charColor[0])*blendFactor)
					blendedG := uint8(float64(baseG)*(1-blendFactor) + float64(charColor[1])*blendFactor)
					blendedB := uint8(float64(baseB)*(1-blendFactor) + float64(charColor[2])*blendFactor)
					graphic.SetPixel(buf, px, py, graphic.Color{blendedR, blendedG, blendedB})
				}
			}
		}
	}
}

// initImageBlocks selects 3x5 block regions from the base image to become falling blocks.
// It prefers blocks from bright areas of the image.
func initImageBlocks(baseRGB []byte) []*imageBlock {
	rng := rand.New(rand.NewSource(blockSeed))

	// Find all valid block positions (top-left corners) with sufficient brightness
	type blockCandidate struct {
		x, y       int
		brightness int
	}

	var candidates []blockCandidate

	// Iterate over possible top-left positions for 3x5 blocks
	for y := 0; y <= graphic.DisplayHeight-blockHeight; y++ {
		for x := 0; x <= graphic.DisplayWidth-blockWidth; x++ {
			// Calculate total brightness of the block
			totalBrightness := 0
			for dy := 0; dy < blockHeight; dy++ {
				for dx := 0; dx < blockWidth; dx++ {
					offset := ((y+dy)*graphic.DisplayWidth + (x + dx)) * 3
					totalBrightness += int(baseRGB[offset]) + int(baseRGB[offset+1]) + int(baseRGB[offset+2])
				}
			}

			// Skip blocks that are too dark (average brightness per pixel < 50)
			avgBrightness := totalBrightness / (blockWidth * blockHeight)
			if avgBrightness < 50 {
				continue
			}

			candidates = append(candidates, blockCandidate{
				x:          x,
				y:          y,
				brightness: totalBrightness,
			})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Shuffle candidates
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// Track selected regions to avoid overlapping blocks
	selected := make(map[int]bool)

	var blocks []*imageBlock

	for _, c := range candidates {
		if len(blocks) >= blockCount {
			break
		}

		// Check if this block overlaps with any already selected block
		overlaps := false
		for dy := 0; dy < blockHeight && !overlaps; dy++ {
			for dx := 0; dx < blockWidth && !overlaps; dx++ {
				key := (c.y+dy)*graphic.DisplayWidth + (c.x + dx)
				if selected[key] {
					overlaps = true
				}
			}
		}
		if overlaps {
			continue
		}

		// Mark all pixels in this block as selected
		for dy := 0; dy < blockHeight; dy++ {
			for dx := 0; dx < blockWidth; dx++ {
				key := (c.y+dy)*graphic.DisplayWidth + (c.x + dx)
				selected[key] = true
			}
		}

		// Extract pixel colors for this block
		var pixels [blockHeight][blockWidth]graphic.Color
		for dy := 0; dy < blockHeight; dy++ {
			for dx := 0; dx < blockWidth; dx++ {
				offset := ((c.y+dy)*graphic.DisplayWidth + (c.x + dx)) * 3
				pixels[dy][dx] = graphic.Color{baseRGB[offset], baseRGB[offset+1], baseRGB[offset+2]}
			}
		}

		// Random start frame offset (0 to matrixFrameCount-1) for staggered falling
		startFrame := rng.Intn(matrixFrameCount)
		speed := 0.5 + rng.Float64()*1.0 // Speed between 0.5 and 1.5 pixels per frame

		// Calculate how many frames until the block exits the screen
		// Block exits when its top edge reaches DisplayHeight
		fallDistance := graphic.DisplayHeight - c.y + blockHeight
		fallFrames := int(float64(fallDistance) / speed)
		if fallFrames >= matrixFrameCount {
			fallFrames = matrixFrameCount - 1
		}

		blocks = append(blocks, &imageBlock{
			srcX:       c.x,
			srcY:       c.y,
			pixels:     pixels,
			startFrame: startFrame,
			speed:      speed,
			fallFrames: fallFrames,
		})
	}

	return blocks
}

// getSourceBrightness calculates the brightness multiplier for a block's source region
func getSourceBrightness(effectiveFrame, fallFrames int) float64 {
	if effectiveFrame < fallFrames {
		// Falling phase: source is dimmed
		return 0.25
	}
	// Restore phase: linear fade from 0.25 to 1.0
	restoreFrames := matrixFrameCount - fallFrames
	if restoreFrames <= 0 {
		return 1.0
	}
	restoreProgress := float64(effectiveFrame-fallFrames) / float64(restoreFrames)
	return 0.25 + 0.75*restoreProgress
}

// drawImageBlocks draws falling blocks and applies source dimming for the current frame
func drawImageBlocks(buf, baseRGB []byte, blocks []*imageBlock, frame int) {
	for _, b := range blocks {
		// Calculate effective frame in this block's cycle
		effectiveFrame := (frame - b.startFrame + matrixFrameCount) % matrixFrameCount

		// Apply source brightness (dimming/restoring the source region)
		brightness := getSourceBrightness(effectiveFrame, b.fallFrames)
		for dy := 0; dy < blockHeight; dy++ {
			for dx := 0; dx < blockWidth; dx++ {
				px, py := b.srcX+dx, b.srcY+dy
				if px >= 0 && px < graphic.DisplayWidth && py >= 0 && py < graphic.DisplayHeight {
					offset := (py*graphic.DisplayWidth + px) * 3
					buf[offset] = uint8(float64(baseRGB[offset]) * brightness)
					buf[offset+1] = uint8(float64(baseRGB[offset+1]) * brightness)
					buf[offset+2] = uint8(float64(baseRGB[offset+2]) * brightness)
				}
			}
		}

		// Draw the falling block if still visible
		screenY := b.srcY + int(float64(effectiveFrame)*b.speed)
		if screenY < graphic.DisplayHeight {
			for dy := 0; dy < blockHeight; dy++ {
				for dx := 0; dx < blockWidth; dx++ {
					py := screenY + dy
					if py >= 0 && py < graphic.DisplayHeight {
						px := b.srcX + dx
						if px >= 0 && px < graphic.DisplayWidth {
							graphic.SetPixel(buf, px, py, b.pixels[dy][dx])
						}
					}
				}
			}
		}
	}
}

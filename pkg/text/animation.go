package text

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// AnimationOptions configures animated text generation.
type AnimationOptions struct {
	TextOptions
	FrameDelay    int // Delay per frame (10ms units, default: 50 = 500ms)
	BlinkOffDelay int // Off-frame delay for blink (default: 30 = 300ms)
	LetterDelay   int // Delay between letters for appear animations (default: 20 = 200ms)
	HoldDelay     int // Hold on final frame (default: 100 = 1s)
}

// DefaultAnimationOptions returns sensible default animation options.
func DefaultAnimationOptions() AnimationOptions {
	return AnimationOptions{
		TextOptions:   DefaultTextOptions(),
		FrameDelay:    50,  // 500ms
		BlinkOffDelay: 30,  // 300ms
		LetterDelay:   20,  // 200ms
		HoldDelay:     100, // 1s
	}
}

// rgbToPaletted converts an RGB buffer to a paletted image for GIF encoding.
func rgbToPaletted(rgbBuf []byte) *image.Paletted {
	rgba := image.NewRGBA(image.Rect(0, 0, graphic.DisplayWidth, graphic.DisplayHeight))
	for y := 0; y < graphic.DisplayHeight; y++ {
		for x := 0; x < graphic.DisplayWidth; x++ {
			offset := (y*graphic.DisplayWidth + x) * 3
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

// GenerateBlinkingText creates a blinking text animation.
// Frame 1: Text ON (centered, shadowed)
// Frame 2: Background only
// LoopCount = 0 (loops forever)
// Automatically wraps text to multiple lines if it doesn't fit.
func GenerateBlinkingText(text string, opts AnimationOptions) *graphic.Image {
	// Frame 1: Text on
	onBuf := graphic.NewBufferWithColor(opts.Background)
	lines := WrapText(text)
	if len(lines) <= 1 {
		DrawTextCentered(onBuf, text, opts.TextOptions)
	} else {
		DrawMultiLineCentered(onBuf, lines, opts.TextOptions)
	}

	// Frame 2: Background only
	offBuf := graphic.NewBufferWithColor(opts.Background)

	g := &gif.GIF{
		Image:     []*image.Paletted{rgbToPaletted(onBuf), rgbToPaletted(offBuf)},
		Delay:     []int{opts.FrameDelay, opts.BlinkOffDelay},
		LoopCount: 0, // Loop forever
	}

	return &graphic.Image{
		Type:    graphic.ImageTypeAnimated,
		GIFData: g,
	}
}

// AppearingFrame represents a single frame in an appearing text animation.
type AppearingFrame struct {
	Data  []byte // Raw RGB buffer
	Delay int    // Delay in 10ms units
}

// GenerateAppearingFrames creates frames for letter-by-letter appearing text.
// Returns individual frames that can be played back in real-time using SendImage.
// This avoids GIF loop issues by giving the caller control over playback.
// Automatically wraps text to multiple lines if it doesn't fit.
func GenerateAppearingFrames(text string, opts AnimationOptions) []AppearingFrame {
	lines := WrapText(text)
	if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
		buf := graphic.NewBufferWithColor(opts.Background)
		return []AppearingFrame{{Data: buf, Delay: opts.HoldDelay}}
	}

	// Build a flat list of (lineIndex, charIndex) for each character
	type charPos struct {
		lineIdx int
		charIdx int
	}
	var positions []charPos
	for lineIdx, line := range lines {
		for charIdx := range []rune(line) {
			positions = append(positions, charPos{lineIdx, charIdx})
		}
	}

	if len(positions) == 0 {
		buf := graphic.NewBufferWithColor(opts.Background)
		return []AppearingFrame{{Data: buf, Delay: opts.HoldDelay}}
	}

	// Calculate vertical centering
	totalHeight := TextBlockHeight(lines)
	startY := (graphic.DisplayHeight - totalHeight) / 2

	var frames []AppearingFrame

	// Frame 0: Background only
	bgBuf := graphic.NewBufferWithColor(opts.Background)
	frames = append(frames, AppearingFrame{Data: bgBuf, Delay: opts.LetterDelay})

	// Frames 1..N: progressively more letters across all lines
	for i := 1; i <= len(positions); i++ {
		buf := graphic.NewBufferWithColor(opts.Background)

		// Determine how many characters to show on each line
		charCount := i
		for lineIdx, line := range lines {
			lineRunes := []rune(line)
			if charCount <= 0 {
				break
			}
			showCount := charCount
			if showCount > len(lineRunes) {
				showCount = len(lineRunes)
			}
			if showCount > 0 {
				partial := string(lineRunes[:showCount])
				lineWidth := TextWidth(line) // Full line width for consistent positioning
				x := (graphic.DisplayWidth - lineWidth) / 2
				y := startY + lineIdx*(FontHeight+LineSpacing)
				DrawTextShadowed(buf, partial, x, y, opts.TextOptions)
			}
			charCount -= len(lineRunes)
		}

		delay := opts.LetterDelay
		if i == len(positions) {
			delay = opts.HoldDelay // Final frame holds longer
		}
		frames = append(frames, AppearingFrame{Data: buf, Delay: delay})
	}

	return frames
}

// GenerateAppearingText creates a letter-by-letter appearing text animation.
// Frames: Background, "H", "HE", "HEL", ... until full text, then hold.
// The final frame has a very long delay (max uint16 = ~10 minutes) to simulate non-looping.
// Automatically wraps text to multiple lines if it doesn't fit.
func GenerateAppearingText(text string, opts AnimationOptions) *graphic.Image {
	lines := WrapText(text)
	if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
		// Empty text: just return a single background frame
		buf := graphic.NewBufferWithColor(opts.Background)
		return &graphic.Image{
			Type: graphic.ImageTypeAnimated,
			GIFData: &gif.GIF{
				Image:     []*image.Paletted{rgbToPaletted(buf)},
				Delay:     []int{65535}, // Max delay (~10 min)
				LoopCount: 0,
			},
		}
	}

	// Count total characters across all lines
	totalChars := 0
	for _, line := range lines {
		totalChars += len([]rune(line))
	}

	if totalChars == 0 {
		buf := graphic.NewBufferWithColor(opts.Background)
		return &graphic.Image{
			Type: graphic.ImageTypeAnimated,
			GIFData: &gif.GIF{
				Image:     []*image.Paletted{rgbToPaletted(buf)},
				Delay:     []int{65535},
				LoopCount: 0,
			},
		}
	}

	// Calculate vertical centering
	totalHeight := TextBlockHeight(lines)
	startY := (graphic.DisplayHeight - totalHeight) / 2

	var frames []*image.Paletted
	var delays []int

	// Frame 0: Background only
	bgBuf := graphic.NewBufferWithColor(opts.Background)
	frames = append(frames, rgbToPaletted(bgBuf))
	delays = append(delays, opts.LetterDelay)

	// Frames 1..N: progressively more letters across all lines
	for i := 1; i <= totalChars; i++ {
		buf := graphic.NewBufferWithColor(opts.Background)

		// Determine how many characters to show on each line
		charCount := i
		for lineIdx, line := range lines {
			lineRunes := []rune(line)
			if charCount <= 0 {
				break
			}
			showCount := charCount
			if showCount > len(lineRunes) {
				showCount = len(lineRunes)
			}
			if showCount > 0 {
				partial := string(lineRunes[:showCount])
				lineWidth := TextWidth(line) // Full line width for consistent positioning
				x := (graphic.DisplayWidth - lineWidth) / 2
				y := startY + lineIdx*(FontHeight+LineSpacing)
				DrawTextShadowed(buf, partial, x, y, opts.TextOptions)
			}
			charCount -= len(lineRunes)
		}

		frames = append(frames, rgbToPaletted(buf))
		if i == totalChars {
			// Final frame: max delay to simulate non-looping (~10 minutes)
			delays = append(delays, 65535)
		} else {
			delays = append(delays, opts.LetterDelay)
		}
	}

	return &graphic.Image{
		Type: graphic.ImageTypeAnimated,
		GIFData: &gif.GIF{
			Image:     frames,
			Delay:     delays,
			LoopCount: 0,
		},
	}
}

// GenerateAppearDisappearText creates a looping animation where text appears
// letter-by-letter, holds, then disappears letter-by-letter (first-to-last removal).
// LoopCount = 0 (loops forever)
// Automatically wraps text to multiple lines if it doesn't fit.
func GenerateAppearDisappearText(text string, opts AnimationOptions) *graphic.Image {
	lines := WrapText(text)
	if len(lines) == 0 || (len(lines) == 1 && len(lines[0]) == 0) {
		buf := graphic.NewBufferWithColor(opts.Background)
		return &graphic.Image{
			Type: graphic.ImageTypeAnimated,
			GIFData: &gif.GIF{
				Image:     []*image.Paletted{rgbToPaletted(buf)},
				Delay:     []int{opts.HoldDelay},
				LoopCount: 0,
			},
		}
	}

	// Count total characters across all lines
	totalChars := 0
	for _, line := range lines {
		totalChars += len([]rune(line))
	}

	if totalChars == 0 {
		buf := graphic.NewBufferWithColor(opts.Background)
		return &graphic.Image{
			Type: graphic.ImageTypeAnimated,
			GIFData: &gif.GIF{
				Image:     []*image.Paletted{rgbToPaletted(buf)},
				Delay:     []int{opts.HoldDelay},
				LoopCount: 0,
			},
		}
	}

	// Calculate vertical centering
	totalHeight := TextBlockHeight(lines)
	startY := (graphic.DisplayHeight - totalHeight) / 2

	var frames []*image.Paletted
	var delays []int

	// Appear sequence: "", "H", "HE", "HEL", ..., full text
	for i := 0; i <= totalChars; i++ {
		buf := graphic.NewBufferWithColor(opts.Background)

		if i > 0 {
			charCount := i
			for lineIdx, line := range lines {
				lineRunes := []rune(line)
				if charCount <= 0 {
					break
				}
				showCount := charCount
				if showCount > len(lineRunes) {
					showCount = len(lineRunes)
				}
				if showCount > 0 {
					partial := string(lineRunes[:showCount])
					lineWidth := TextWidth(line)
					x := (graphic.DisplayWidth - lineWidth) / 2
					y := startY + lineIdx*(FontHeight+LineSpacing)
					DrawTextShadowed(buf, partial, x, y, opts.TextOptions)
				}
				charCount -= len(lineRunes)
			}
		}

		frames = append(frames, rgbToPaletted(buf))
		if i == totalChars {
			delays = append(delays, opts.HoldDelay) // Hold on full text
		} else {
			delays = append(delays, opts.LetterDelay)
		}
	}

	// Disappear sequence (first-to-last removal across all lines)
	for i := 1; i <= totalChars; i++ {
		buf := graphic.NewBufferWithColor(opts.Background)

		if i < totalChars {
			// Skip first i characters across all lines
			skipCount := i
			for lineIdx, line := range lines {
				lineRunes := []rune(line)
				if skipCount >= len(lineRunes) {
					// Skip entire line
					skipCount -= len(lineRunes)
					continue
				}
				// Show remaining characters on this line
				remaining := string(lineRunes[skipCount:])
				lineWidth := TextWidth(line)
				x := (graphic.DisplayWidth - lineWidth) / 2
				// Shift x position to account for removed characters
				xOffset := x + skipCount*FontSpacing
				y := startY + lineIdx*(FontHeight+LineSpacing)
				DrawTextShadowed(buf, remaining, xOffset, y, opts.TextOptions)
				skipCount = 0 // Remaining lines show in full

				// Show subsequent lines in full
				for nextLineIdx := lineIdx + 1; nextLineIdx < len(lines); nextLineIdx++ {
					nextLine := lines[nextLineIdx]
					if len(nextLine) == 0 {
						continue
					}
					nextLineWidth := TextWidth(nextLine)
					nextX := (graphic.DisplayWidth - nextLineWidth) / 2
					nextY := startY + nextLineIdx*(FontHeight+LineSpacing)
					DrawTextShadowed(buf, nextLine, nextX, nextY, opts.TextOptions)
				}
				break
			}
		}

		frames = append(frames, rgbToPaletted(buf))
		if i == totalChars {
			delays = append(delays, opts.HoldDelay) // Hold on empty
		} else {
			delays = append(delays, opts.LetterDelay)
		}
	}

	return &graphic.Image{
		Type: graphic.ImageTypeAnimated,
		GIFData: &gif.GIF{
			Image:     frames,
			Delay:     delays,
			LoopCount: 0, // Loop forever
		},
	}
}

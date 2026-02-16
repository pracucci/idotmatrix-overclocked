package text

import (
	"strings"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// TextOptions configures text rendering.
type TextOptions struct {
	TextColor   graphic.Color // Main text color
	ShadowColor graphic.Color // Shadow color (use Black with ShadowX/Y=0 to disable)
	Background  graphic.Color // Background fill color
	ShadowX     int       // Shadow X offset (default: 1)
	ShadowY     int       // Shadow Y offset (default: 1)
}

// DefaultTextOptions returns sensible default options.
func DefaultTextOptions() TextOptions {
	return TextOptions{
		TextColor:   graphic.White,
		ShadowColor: graphic.Gray,
		Background:  graphic.Black,
		ShadowX:     1,
		ShadowY:     1,
	}
}

// DrawTextShadowed draws text with a shadow at the given position.
// Returns the width of the drawn text.
func DrawTextShadowed(buf []byte, text string, x, y int, opts TextOptions) int {
	// Draw shadow first (if offset is non-zero)
	if opts.ShadowX != 0 || opts.ShadowY != 0 {
		DrawText(buf, text, x+opts.ShadowX, y+opts.ShadowY, opts.ShadowColor)
	}
	// Draw main text
	return DrawText(buf, text, x, y, opts.TextColor)
}

// DrawTextCentered draws single-line text centered on the display with shadow.
// Returns the calculated x, y position where the text was drawn.
func DrawTextCentered(buf []byte, text string, opts TextOptions) (x, y int) {
	textW := TextWidth(text)
	x = (graphic.DisplayWidth - textW) / 2
	y = (graphic.DisplayHeight - FontHeight) / 2
	DrawTextShadowed(buf, text, x, y, opts)
	return x, y
}

// WrapText wraps text to fit within the display width.
// Words are kept together when possible; long words are broken character by character.
func WrapText(text string) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		wordWidth := TextWidth(word)

		// If word itself is too wide, break it character by character
		if wordWidth > graphic.DisplayWidth {
			// First, flush current line if not empty
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			// Break the word into fitting chunks
			runes := []rune(word)
			chunk := ""
			for _, r := range runes {
				testChunk := chunk + string(r)
				if TextWidth(testChunk) > graphic.DisplayWidth {
					if chunk != "" {
						lines = append(lines, chunk)
					}
					chunk = string(r)
				} else {
					chunk = testChunk
				}
			}
			if chunk != "" {
				currentLine = chunk
			}
			continue
		}

		if currentLine == "" {
			currentLine = word
		} else {
			// Try adding word to current line with a space
			testLine := currentLine + " " + word
			if TextWidth(testLine) <= graphic.DisplayWidth {
				currentLine = testLine
			} else {
				// Start new line
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
	}

	// Don't forget the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// TextBlockHeight calculates total height of multi-line text block.
func TextBlockHeight(lines []string) int {
	if len(lines) == 0 {
		return 0
	}
	return len(lines)*FontHeight + (len(lines)-1)*LineSpacing
}

// DrawMultiLineCentered draws multi-line text centered both horizontally and vertically.
func DrawMultiLineCentered(buf []byte, lines []string, opts TextOptions) {
	if len(lines) == 0 {
		return
	}

	totalHeight := TextBlockHeight(lines)
	startY := (graphic.DisplayHeight - totalHeight) / 2

	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		lineWidth := TextWidth(line)
		x := (graphic.DisplayWidth - lineWidth) / 2
		y := startY + i*(FontHeight+LineSpacing)
		DrawTextShadowed(buf, line, x, y, opts)
	}
}

// GenerateStaticText creates a static image with centered, shadowed text.
// Automatically wraps text to multiple lines if it doesn't fit.
func GenerateStaticText(text string, opts TextOptions) *graphic.Image {
	buf := graphic.NewBufferWithColor(opts.Background)
	lines := WrapText(text)
	if len(lines) <= 1 {
		DrawTextCentered(buf, text, opts)
	} else {
		DrawMultiLineCentered(buf, lines, opts)
	}
	return &graphic.Image{
		Type:       graphic.ImageTypeStatic,
		StaticData: buf,
	}
}

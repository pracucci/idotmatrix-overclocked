package text

import "github.com/pracucci/idotmatrix-overclocked/pkg/graphic"

// DrawChar draws a single character at the given position using the 5x7 font.
// Returns the width of the character drawn (FontWidth for known chars, 0 for unknown).
func DrawChar(buf []byte, char rune, x, y int, color graphic.Color) int {
	data, ok := font5x7[char]
	if !ok {
		return 0
	}
	for row := 0; row < FontHeight; row++ {
		for col := 0; col < FontWidth; col++ {
			if data[row]&(1<<col) != 0 {
				graphic.SetPixel(buf, x+col, y+row, color)
			}
		}
	}
	return FontWidth
}

// DrawText draws a string of text at the given position.
// Returns the total width in pixels of the drawn text.
func DrawText(buf []byte, text string, x, y int, color graphic.Color) int {
	startX := x
	for _, char := range text {
		DrawChar(buf, char, x, y, color)
		x += FontSpacing
	}
	return x - startX - 1 // Subtract trailing gap
}

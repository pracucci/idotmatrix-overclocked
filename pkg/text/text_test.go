package text

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

func TestTextWidth(t *testing.T) {
	t.Run("empty string returns 0", func(t *testing.T) {
		w := TextWidth("")
		assert.Equal(t, 0, w)
	})

	t.Run("single character", func(t *testing.T) {
		// len("A") * FontSpacing - 1 = 1 * 6 - 1 = 5
		w := TextWidth("A")
		expected := 1*FontSpacing - 1
		assert.Equal(t, expected, w)
	})

	t.Run("multiple characters", func(t *testing.T) {
		// len("AB") * FontSpacing - 1 = 2 * 6 - 1 = 11
		w := TextWidth("AB")
		expected := 2*FontSpacing - 1
		assert.Equal(t, expected, w)
	})
}

func TestWrapText(t *testing.T) {
	t.Run("empty string returns empty slice", func(t *testing.T) {
		lines := WrapText("")
		assert.Empty(t, lines)
	})

	t.Run("single short word returns single-element slice", func(t *testing.T) {
		lines := WrapText("HI")
		require.Len(t, lines, 1)
		assert.Equal(t, "HI", lines[0])
	})

	t.Run("multiple words that fit on one line stay together", func(t *testing.T) {
		// "A B" has width: 3 chars * 6 - 1 = 17 pixels (fits in 64)
		lines := WrapText("A B")
		require.Len(t, lines, 1)
		assert.Equal(t, "A B", lines[0])
	})

	t.Run("words that exceed width wrap to next line", func(t *testing.T) {
		// Create two words that together exceed 64px but individually fit
		// Each word with 8 chars: 8 * 6 - 1 = 47px (fits)
		// Together with space: 17 chars * 6 - 1 = 101px (doesn't fit)
		word1 := "AAAAAAAA"
		word2 := "BBBBBBBB"
		lines := WrapText(word1 + " " + word2)

		require.Len(t, lines, 2)
		assert.Equal(t, word1, lines[0])
		assert.Equal(t, word2, lines[1])
	})

	t.Run("very long word breaks character by character", func(t *testing.T) {
		// Create a word that's > 64px wide
		// 64px / 6px per char ~= 10.6, so 11 chars = 65px which exceeds 64
		longWord := strings.Repeat("X", 15) // 15 * 6 - 1 = 89px
		lines := WrapText(longWord)

		require.GreaterOrEqual(t, len(lines), 2, "should break long word into multiple lines")

		// Verify all characters are accounted for
		var totalChars int
		for _, line := range lines {
			totalChars += len(line)
		}
		assert.Equal(t, 15, totalChars)
	})
}

func TestTextBlockHeight(t *testing.T) {
	t.Run("empty slice returns 0", func(t *testing.T) {
		h := TextBlockHeight([]string{})
		assert.Equal(t, 0, h)
	})

	t.Run("single line returns FontHeight", func(t *testing.T) {
		h := TextBlockHeight([]string{"A"})
		assert.Equal(t, FontHeight, h)
	})

	t.Run("two lines includes line spacing", func(t *testing.T) {
		// 2*FontHeight + LineSpacing = 2*7 + 4 = 18
		h := TextBlockHeight([]string{"A", "B"})
		expected := 2*FontHeight + LineSpacing
		assert.Equal(t, expected, h)
	})
}

func TestDrawChar(t *testing.T) {
	t.Run("known character returns FontWidth", func(t *testing.T) {
		buf := graphic.NewBuffer()
		w := DrawChar(buf, 'A', 0, 0, graphic.White)
		assert.Equal(t, FontWidth, w)
	})

	t.Run("unknown character returns 0", func(t *testing.T) {
		buf := graphic.NewBuffer()
		w := DrawChar(buf, '\u00A9', 0, 0, graphic.White) // copyright symbol
		assert.Equal(t, 0, w)
	})
}

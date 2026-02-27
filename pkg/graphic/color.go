package graphic

// Color represents an RGB color.
type Color [3]uint8

// Predefined colors
var (
	Black   = Color{0, 0, 0}
	White   = Color{255, 255, 255}
	Red     = Color{255, 0, 0}
	Green   = Color{0, 255, 0}
	Blue    = Color{0, 0, 255}
	Yellow  = Color{255, 255, 0}
	Cyan    = Color{0, 255, 255}
	Magenta = Color{255, 0, 255}
	Orange  = Color{255, 150, 0}
	Gray    = Color{128, 128, 128}
	DimGray = Color{80, 80, 80}
	Purple  = Color{128, 0, 128}
	Violet  = Color{160, 32, 240}
	Pink    = Color{255, 105, 180}

	// Shadow colors (darker versions)
	DarkRed     = Color{50, 0, 0}
	DarkGreen   = Color{0, 50, 0}
	DarkBlue    = Color{0, 0, 50}
	DarkYellow  = Color{50, 50, 0}
	DarkCyan    = Color{0, 50, 50}
	DarkMagenta = Color{50, 0, 50}
	DarkOrange  = Color{50, 30, 0}
	DarkGray    = Color{30, 30, 30}
	DarkPurple  = Color{50, 20, 35}
	DarkPink    = Color{50, 20, 35}
	DarkWhite   = Color{40, 40, 40}
)

// ColorPalette maps color names to Color values.
var ColorPalette = map[string]Color{
	"black":   Black,
	"white":   White,
	"red":     Red,
	"green":   Green,
	"blue":    Blue,
	"yellow":  Yellow,
	"cyan":    Cyan,
	"magenta": Magenta,
	"orange":  Orange,
	"gray":    Gray,
	"grey":    Gray,
	"purple":  Purple,
	"pink":    Pink,
}

// ColorShadows maps each color to its shadow (darker) version.
var ColorShadows = map[Color]Color{
	White:   DarkWhite,
	Red:     DarkRed,
	Green:   DarkGreen,
	Blue:    DarkBlue,
	Yellow:  DarkYellow,
	Cyan:    DarkCyan,
	Magenta: DarkMagenta,
	Orange:  DarkOrange,
	Gray:    DarkGray,
	Purple:  DarkPurple,
	Pink:    DarkPink,
	Black:   Black, // Black has no shadow
}

// ShadowFor returns the shadow color for a given color.
func ShadowFor(c Color) Color {
	if shadow, ok := ColorShadows[c]; ok {
		return shadow
	}
	// Fallback: create shadow by dividing RGB by 5
	return Color{c[0] / 5, c[1] / 5, c[2] / 5}
}

// ColorNames returns a list of available color names.
func ColorNames() []string {
	return []string{"white", "red", "green", "blue", "yellow", "cyan", "magenta", "orange", "gray", "purple", "pink"}
}

package text

import (
	"strings"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// AnimationType represents a supported text animation.
type AnimationType struct {
	Name        string // Primary name
	Description string // Human-readable description
}

// AnimationTypes defines all supported text animations.
var AnimationTypes = []AnimationType{
	{
		Name:        "none",
		Description: "Static centered text (default)",
	},
	{
		Name:        "blink",
		Description: "Text blinks on/off (loops forever)",
	},
	{
		Name:        "appear",
		Description: "Letters appear one by one (plays once)",
	},
	{
		Name:        "appear-disappear",
		Description: "Letters appear then disappear (loops forever)",
	},
	{
		Name:        "fireworks",
		Description: "Text with colorful fireworks (loops forever)",
	},
}

// AnimationTypeNames returns a list of primary animation type names.
func AnimationTypeNames() []string {
	names := make([]string, len(AnimationTypes))
	for i, at := range AnimationTypes {
		names[i] = at.Name
	}
	return names
}

// AnimationTypeNamesString returns animation type names as a comma-separated string.
func AnimationTypeNamesString() string {
	return strings.Join(AnimationTypeNames(), ", ")
}

// GenerateAnimation generates the appropriate animation based on type name.
// Returns nil and an error message if the animation type is unknown.
func GenerateAnimation(animationType, text string, opts AnimationOptions) (*graphic.Image, string) {
	switch strings.ToLower(strings.TrimSpace(animationType)) {
	case "none":
		return GenerateStaticText(text, opts.TextOptions), ""
	case "blink":
		return GenerateBlinkingText(text, opts), ""
	case "appear":
		return GenerateAppearingText(text, opts), ""
	case "appear-disappear":
		return GenerateAppearDisappearText(text, opts), ""
	case "fireworks":
		return GenerateFireworksText(text, opts), ""
	default:
		return nil, "unknown animation type: " + animationType + " (valid: " + AnimationTypeNamesString() + ")"
	}
}

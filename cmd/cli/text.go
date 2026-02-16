package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/pracucci/idotmatrix-overclocked/pkg/text"
	"github.com/spf13/cobra"
)

var (
	textTargetAddr string
	textMsg        string
	textAnimation  string
	textColorName  string
	textVerbose    bool
)

var TextCmd = &cobra.Command{
	Use:   "text",
	Short: "Displays text on the 64x64 iDot display",
	Long: `Displays text on the 64x64 iDot display with optional animation.

Text is automatically word-wrapped to fit the display width.

Animation options:
  none              - Static centered text (default)
  blink             - Text blinks on/off (loops forever)
  appear            - Letters appear one by one (plays once)
  appear-disappear  - Letters appear then disappear (loops forever)

Color options: ` + strings.Join(graphic.ColorNames(), ", ") + `

Examples:
  idm-cli text --target AA:BB:CC:DD:EE:FF --text "HELLO"
  idm-cli text --target AA:BB:CC:DD:EE:FF --text "HELLO WORLD"
  idm-cli text --target AA:BB:CC:DD:EE:FF --text "HI" --animation blink
  idm-cli text --target AA:BB:CC:DD:EE:FF --text "HELLO" --color red`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(textVerbose)
		if err := doShowText(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	TextCmd.Flags().StringVar(&textTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")

	TextCmd.Flags().StringVar(&textMsg, "text", "", "Text to display (uppercase A-Z, 0-9, space, basic punctuation)")
	TextCmd.MarkFlagRequired("text")

	TextCmd.Flags().StringVar(&textAnimation, "animation", "none", "Animation type: none, blink, appear, appear-disappear")
	TextCmd.Flags().StringVar(&textColorName, "color", "white", fmt.Sprintf("Text color (%s)", strings.Join(graphic.ColorNames(), ", ")))
	TextCmd.Flags().BoolVar(&textVerbose, "verbose", false, "Enable verbose debug logging")
}

func doShowText(logger log.Logger) error {
	if len(textMsg) == 0 {
		return fmt.Errorf("missing --text option")
	}

	// Normalize animation flag
	animation := strings.ToLower(strings.TrimSpace(textAnimation))

	// Convert text to uppercase (font only has uppercase)
	msg := strings.ToUpper(textMsg)

	// Wrap text and validate total height fits
	lines := text.WrapText(msg)
	blockHeight := text.TextBlockHeight(lines)
	if blockHeight > graphic.DisplayHeight {
		return fmt.Errorf("text too long: wrapped to %d lines (%d pixels, max %d)", len(lines), blockHeight, graphic.DisplayHeight)
	}

	// Parse color
	colorName := strings.ToLower(strings.TrimSpace(textColorName))
	color, ok := graphic.ColorPalette[colorName]
	if !ok {
		return fmt.Errorf("unknown color: %s (valid: %s)", colorName, strings.Join(graphic.ColorNames(), ", "))
	}

	// Generate the image based on animation type
	opts := text.DefaultAnimationOptions()
	opts.TextOptions.TextColor = color
	opts.TextOptions.ShadowColor = graphic.ShadowFor(color)
	var image *graphic.Image

	switch animation {
	case "none", "static":
		image = text.GenerateStaticText(msg, opts.TextOptions)
	case "blink":
		image = text.GenerateBlinkingText(msg, opts)
	case "appear":
		image = text.GenerateAppearingText(msg, opts)
	case "appear-disappear", "appear-and-disappear":
		image = text.GenerateAppearDisappearText(msg, opts)
	default:
		return fmt.Errorf("unknown animation type: %s (valid: none, blink, appear, appear-disappear)", animation)
	}

	// Connect to device
	device := protocol.NewDevice(logger)
	if err := device.Connect(textTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	// Send to device based on image type
	if image.Type == graphic.ImageTypeStatic {
		if err := protocol.SetDrawMode(device, 1); err != nil {
			return err
		}
		rawBytes, err := image.RawBytes()
		if err != nil {
			return err
		}
		if err := protocol.SendImage(device, rawBytes); err != nil {
			return err
		}
	} else {
		gifBytes, err := image.GIFBytes()
		if err != nil {
			return err
		}
		if err := protocol.SendGIF(device, gifBytes, logger); err != nil {
			return err
		}
	}

	// Allow time for BLE writes to complete before disconnecting
	time.Sleep(500 * time.Millisecond)

	return nil
}

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/assets"
	"github.com/pracucci/idotmatrix-overclocked/pkg/emoji"
	"github.com/pracucci/idotmatrix-overclocked/pkg/fire"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/grot"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/pracucci/idotmatrix-overclocked/pkg/text"
	"github.com/spf13/cobra"
)

var (
	demoTargetAddr string
	demoVerbose    bool
)

var DemoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Showcase all display features in a slideshow",
	Long: `Showcase all non-interactive display features in a continuous slideshow.

The demo cycles through emojis, animations, and text effects in random order.
Press Ctrl+C to stop.

Examples:
  idm-cli demo
  idm-cli demo --target AA:BB:CC:DD:EE:FF`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(demoVerbose)
		if err := doDemo(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	DemoCmd.Flags().StringVar(&demoTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	DemoCmd.Flags().BoolVar(&demoVerbose, "verbose", false, "Enable verbose debug logging")
}

// demoItem represents one item in the demo sequence
type demoItem struct {
	name     string
	generate func() ([]byte, error)
	duration time.Duration
}

func doDemo(logger log.Logger) error {
	// Connect to device
	device := protocol.NewDevice(logger)
	if err := device.Connect(demoTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	items := []demoItem{
		{"snake", func() ([]byte, error) { return assets.Preview.ReadFile("preview/snake-preview.gif") }, 4 * time.Second},
		{"emoji: rocket", generateEmojiGIF("rocket"), 3 * time.Second},
		{"emoji: thumbsup", generateEmojiGIF("thumbsup"), 3 * time.Second},
		{"emoji: rofl", generateEmojiGIF("rofl"), 3 * time.Second},
		{"grot: halloween-4", generateGrotGIF("halloween-4"), 3 * time.Second},
		{"grot: matrix", generateGrotGIF("matrix"), 4 * time.Second},
		{"fire", generateFireGIF(), 3 * time.Second},
		{"text: FIRE! (fireworks)", generateTextGIF("FIRE!", "fireworks", graphic.Red), 4 * time.Second},
		{"text: LGTM (appear-disappear)", generateTextGIF("LGTM", "appear-disappear", graphic.Green), 4 * time.Second},
	}

	fmt.Printf("Starting demo with %d items (random order)\n", len(items))
	fmt.Println("Press Ctrl+C to stop")

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		// Shuffle items for each iteration
		rng.Shuffle(len(items), func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})

		for _, item := range items {
			fmt.Printf("Showing: %s\n", item.name)

			gifBytes, err := item.generate()
			if err != nil {
				level.Error(logger).Log("msg", "Failed to generate", "name", item.name, "err", err)
				continue
			}

			if err := protocol.SendGIF(device, gifBytes, logger); err != nil {
				level.Error(logger).Log("msg", "Failed to send GIF", "name", item.name, "err", err)
				continue
			}

			time.Sleep(item.duration)
		}
	}
}

func generateEmojiGIF(name string) func() ([]byte, error) {
	return func() ([]byte, error) {
		img, err := emoji.Generate(name)
		if err != nil {
			return nil, err
		}
		return img.GIFBytes()
	}
}

func generateGrotGIF(name string) func() ([]byte, error) {
	return func() ([]byte, error) {
		img, err := grot.Generate(name)
		if err != nil {
			return nil, err
		}
		return img.GIFBytes()
	}
}

func generateFireGIF() func() ([]byte, error) {
	return func() ([]byte, error) {
		return fire.GenerateGIF(), nil
	}
}

func generateTextGIF(msg, animation string, color graphic.Color) func() ([]byte, error) {
	return func() ([]byte, error) {
		opts := text.DefaultAnimationOptions()
		opts.TextColor = color
		opts.ShadowColor = graphic.ShadowFor(color)
		img, errMsg := text.GenerateAnimation(animation, msg, opts)
		if errMsg != "" {
			return nil, fmt.Errorf("%s", errMsg)
		}
		return img.GIFBytes()
	}
}

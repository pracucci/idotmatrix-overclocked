package main

import (
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/emoji"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
)

var (
	emojiSlideshowTargetAddr string
	emojiSlideshowInterval   int
	emojiSlideshowVerbose    bool
)

var EmojiSlideshowCmd = &cobra.Command{
	Use:   "emoji-slideshow",
	Short: "Display all emojis in a slideshow",
	Long: `Display all available emojis one by one in a slideshow on the 64x64 iDot display.

Each emoji is displayed for a configurable interval (default 3 seconds).
The slideshow loops forever until interrupted.

Examples:
  idm-cli emoji-slideshow
  idm-cli emoji-slideshow --interval 5
  idm-cli emoji-slideshow --target AA:BB:CC:DD:EE:FF`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(emojiSlideshowVerbose)
		if err := doEmojiSlideshow(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	EmojiSlideshowCmd.Flags().StringVar(&emojiSlideshowTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	EmojiSlideshowCmd.Flags().IntVar(&emojiSlideshowInterval, "interval", 3, "Seconds to display each emoji")
	EmojiSlideshowCmd.Flags().BoolVar(&emojiSlideshowVerbose, "verbose", false, "Enable verbose debug logging")
}

func doEmojiSlideshow(logger log.Logger) error {
	// Connect to device
	device := protocol.NewDevice(logger)
	if err := device.Connect(emojiSlideshowTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	emojiNames := emoji.PrimaryNames()
	interval := time.Duration(emojiSlideshowInterval) * time.Second

	fmt.Printf("Starting emoji slideshow with %d emojis (%ds interval)\n", len(emojiNames), emojiSlideshowInterval)
	fmt.Println("Press Ctrl+C to stop")

	for {
		for _, name := range emojiNames {
			fmt.Printf("Showing: %s\n", name)

			image, err := emoji.Generate(name)
			if err != nil {
				level.Error(logger).Log("msg", "Failed to generate emoji", "name", name, "err", err)
				continue
			}

			gifBytes, err := image.GIFBytes()
			if err != nil {
				level.Error(logger).Log("msg", "Failed to encode GIF", "name", name, "err", err)
				continue
			}

			if err := protocol.SendGIF(device, gifBytes, logger); err != nil {
				level.Error(logger).Log("msg", "Failed to send GIF", "name", name, "err", err)
				continue
			}

			time.Sleep(interval)
		}
	}
}

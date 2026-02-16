package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/emoji"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
)

var (
	emojiTargetAddr string
	emojiName       string
	emojiVerbose    bool
)

var EmojiCmd = &cobra.Command{
	Use:   "emoji",
	Short: "Display an animated emoji on the iDot display",
	Long: fmt.Sprintf(`Display a large animated colorful emoji on the 64x64 iDot display.

Available emojis: %s

Emoji animations from: https://googlefonts.github.io/noto-emoji-animation/

Examples:
  idm-cli emoji --name thumbsup
  idm-cli emoji --name +1
  idm-cli emoji --name party
  idm-cli emoji --target AA:BB:CC:DD:EE:FF --name rocket`, strings.Join(emoji.Names(), ", ")),
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(emojiVerbose)
		if err := doEmoji(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	EmojiCmd.Flags().StringVar(&emojiTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")

	EmojiCmd.Flags().StringVar(&emojiName, "name", "", fmt.Sprintf("Emoji name (%s)", strings.Join(emoji.Names(), ", ")))
	EmojiCmd.MarkFlagRequired("name")

	EmojiCmd.Flags().BoolVar(&emojiVerbose, "verbose", false, "Enable verbose debug logging")
}

func doEmoji(logger log.Logger) error {
	if len(emojiName) == 0 {
		return fmt.Errorf("missing --name option")
	}

	// Generate emoji image
	image, err := emoji.Generate(emojiName)
	if err != nil {
		return err
	}

	// Connect to device
	device := protocol.NewDevice(logger)
	if err := device.Connect(emojiTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	// Send animated GIF to device
	gifBytes, err := image.GIFBytes()
	if err != nil {
		return err
	}

	if err := protocol.SendGIF(device, gifBytes, logger); err != nil {
		return err
	}

	// Allow time for BLE writes to complete before disconnecting
	time.Sleep(500 * time.Millisecond)

	return nil
}

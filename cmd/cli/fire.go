package main

import (
	"bytes"
	"fmt"
	"image/gif"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/fire"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
)

var fireTargetAddr string
var fireVerbose bool
var fireMirrored bool
var fireBrightness int

var FireCmd = &cobra.Command{
	Use:   "fire",
	Short: "Generate and display a DOOM-style fire animation",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(fireVerbose)
		if err := doFire(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	FireCmd.Flags().StringVar(&fireTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	FireCmd.Flags().BoolVar(&fireVerbose, "verbose", false, "Enable verbose debug logging")
	FireCmd.Flags().BoolVar(&fireMirrored, "mirrored", false, "Mirror the image horizontally")
	FireCmd.Flags().IntVar(&fireBrightness, "brightness", 100, "Brightness level (0-100)")
}

func doFire(logger log.Logger) error {
	fmt.Println("Generating DOOM fire animation...")
	gifData := fire.GenerateGIF()

	g, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err != nil {
		return fmt.Errorf("failed to decode GIF: %w", err)
	}
	if fireMirrored {
		g = graphic.MirrorGIFHorizontal(g)
	}
	g = graphic.AdjustBrightnessGIF(g, fireBrightness)
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return fmt.Errorf("failed to re-encode GIF: %w", err)
	}
	gifData = buf.Bytes()

	fmt.Printf("Generated GIF: %d bytes\n", len(gifData))

	device := protocol.NewDevice(logger)
	if err := device.Connect(fireTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	if err := protocol.SendGIF(device, gifData, logger); err != nil {
		return err
	}

	// Allow time for final writes to complete
	time.Sleep(500 * time.Millisecond)

	return nil
}

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/grot"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
)

var (
	grotTargetAddr string
	grotName       string
	grotVerbose    bool
)

var GrotCmd = &cobra.Command{
	Use:   "grot",
	Short: "Display a Halloween-themed animation on the iDot display",
	Long: fmt.Sprintf(`Display a Halloween-themed animated image on the 64x64 iDot display.

Available grots: %s

Grot animations from Giphy.

Examples:
  idm-cli grot --name halloween-1
  idm-cli grot --name halloween-3
  idm-cli grot --target AA:BB:CC:DD:EE:FF --name halloween-5`, strings.Join(grot.Names(), ", ")),
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(grotVerbose)
		if err := doGrot(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	GrotCmd.Flags().StringVar(&grotTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")

	GrotCmd.Flags().StringVar(&grotName, "name", "", fmt.Sprintf("Grot name (%s)", strings.Join(grot.Names(), ", ")))
	GrotCmd.MarkFlagRequired("name")

	GrotCmd.Flags().BoolVar(&grotVerbose, "verbose", false, "Enable verbose debug logging")
}

func doGrot(logger log.Logger) error {
	if len(grotName) == 0 {
		return fmt.Errorf("missing --name option")
	}

	// Generate grot image
	image, err := grot.Generate(grotName)
	if err != nil {
		return err
	}

	// Connect to device
	device := protocol.NewDevice(logger)
	if err := device.Connect(grotTargetAddr); err != nil {
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

	level.Info(logger).Log("msg", "Uploading GIF to device", "name", grotName)
	if err := protocol.SendGIF(device, gifBytes, logger); err != nil {
		return err
	}
	level.Info(logger).Log("msg", "GIF upload complete", "name", grotName)

	// Allow time for BLE writes to complete before disconnecting
	time.Sleep(500 * time.Millisecond)

	return nil
}

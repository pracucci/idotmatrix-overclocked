package main

import (
	"fmt"
	"time"

	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
)

var onTargetAddr string
var onVerbose bool

var OnCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn the iDot display on",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(onVerbose)
		if err := doSetPower(logger, onTargetAddr, true); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

var offTargetAddr string
var offVerbose bool

var OffCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn the iDot display off",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(offVerbose)
		if err := doSetPower(logger, offTargetAddr, false); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	OnCmd.Flags().StringVar(&onTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	OnCmd.Flags().BoolVar(&onVerbose, "verbose", false, "Enable verbose debug logging")

	OffCmd.Flags().StringVar(&offTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	OffCmd.Flags().BoolVar(&offVerbose, "verbose", false, "Enable verbose debug logging")
}

func doSetPower(logger interface{ Log(...interface{}) error }, targetAddr string, on bool) error {
	device := protocol.NewDevice(logger)
	if err := device.Connect(targetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	if err := protocol.SetPowerState(device, on); err != nil {
		return err
	}

	// Allow time for BLE writes to complete before disconnecting
	time.Sleep(500 * time.Millisecond)

	return nil
}

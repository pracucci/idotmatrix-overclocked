package main

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"

	"github.com/pracucci/idotmatrix-overclocked/pkg/games/tetris"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
)

var (
	tetrisTargetAddr string
	tetrisVerbose    bool
)

var TetrisCmd = &cobra.Command{
	Use:   "tetris",
	Short: "Play Tetris on the iDot display",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(tetrisVerbose)
		if err := runTetris(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	TetrisCmd.Flags().StringVar(&tetrisTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	TetrisCmd.Flags().BoolVar(&tetrisVerbose, "verbose", false, "Enable verbose debug logging")
}

func runTetris(logger log.Logger) error {
	device := protocol.NewDevice(logger)
	if err := device.Connect(tetrisTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	game := tetris.NewGame(device)
	return game.Run()
}

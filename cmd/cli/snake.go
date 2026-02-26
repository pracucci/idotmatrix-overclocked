package main

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"

	"github.com/pracucci/idotmatrix-overclocked/pkg/games/snake"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
)

var (
	snakeTargetAddr string
	snakeVerbose    bool
)

var SnakeCmd = &cobra.Command{
	Use:   "snake",
	Short: "Play Snake on the iDot display",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(snakeVerbose)
		if err := runSnake(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	SnakeCmd.Flags().StringVar(&snakeTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	SnakeCmd.Flags().BoolVar(&snakeVerbose, "verbose", false, "Enable verbose debug logging")
}

func runSnake(logger log.Logger) error {
	device := protocol.NewDevice(logger)
	if err := device.Connect(snakeTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	game := snake.NewGame(device, 1)
	return game.Run()
}

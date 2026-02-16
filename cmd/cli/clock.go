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
	"github.com/spf13/cobra"
)

var clockClockStyle int
var clockShowDate bool
var clockShow24h bool
var clockColor string
var clockTimeValue string
var clockTargetAddr string
var clockVerbose bool

var ClockCmd = &cobra.Command{
	Use:   "clock",
	Short: "Shows and optionally configures the clock of the iDot display",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(clockVerbose)
		if err := doSetClock(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	ClockCmd.Flags().StringVar(&clockTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	ClockCmd.Flags().StringVar(&clockTimeValue, "time", "", "Time value in RFC1123Z format. As per 'date -R'")
	ClockCmd.Flags().IntVar(&clockClockStyle, "style", protocol.ClockAnimatedHourGlass, "Style of clock. 0:Default 1:Christmas 2:Racing 3:Inverted 4:Hour Glass")
	ClockCmd.Flags().BoolVar(&clockShowDate, "show-date", true, "Show date as well as time")
	ClockCmd.Flags().BoolVar(&clockShow24h, "24hour", true, "Show time in 24 hour format")
	ClockCmd.Flags().StringVar(&clockColor, "color", "white", fmt.Sprintf("Clock color (%s)", strings.Join(graphic.ColorNames(), ", ")))
	ClockCmd.Flags().BoolVar(&clockVerbose, "verbose", false, "Enable verbose debug logging")
}

func doSetClock(logger log.Logger) error {
	if clockClockStyle > protocol.ClockAnimatedHourGlass {
		return fmt.Errorf("invalid style")

	}

	var t time.Time
	var err error

	if len(clockTimeValue) > 0 {
		t, err = time.Parse(time.RFC1123Z, clockTimeValue)
		if err != nil {
			return err
		}
	} else {
		t = time.Now()
	}

	device := protocol.NewDevice(logger)
	if err = device.Connect(clockTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	if err = protocol.SetTime(device, t.Year(), int(t.Month()), t.Day(), int(t.Weekday())+1, t.Hour(),
		t.Minute(), t.Second()); err != nil {
		return err
	}

	var customColor graphic.Color
	if clockColor != "" {
		colorName := strings.ToLower(strings.TrimSpace(clockColor))
		color, ok := graphic.ColorPalette[colorName]
		if !ok {
			return fmt.Errorf("unknown color: %s (valid: %s)", colorName, strings.Join(graphic.ColorNames(), ", "))
		}
		customColor = graphic.Color{color[0], color[1], color[2]}
	}

	if err := protocol.SetClockMode(device, clockClockStyle, clockShowDate, clockShow24h, customColor); err != nil {
		return err
	}

	// Allow time for BLE writes to complete before disconnecting
	time.Sleep(500 * time.Millisecond)

	return nil
}

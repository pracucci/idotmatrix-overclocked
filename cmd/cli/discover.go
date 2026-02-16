package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/spf13/cobra"
	"tinygo.org/x/bluetooth"
)

var discoverMaxScanTime uint32
var discoverVerbose bool

var DiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover nearby Bluetooth devices",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(discoverVerbose)
		if err := doBTScan(logger); err != nil {
			fmt.Printf("Failed: %v\n", err)
		}
	},
}

func init() {
	DiscoverCmd.Flags().Uint32Var(&discoverMaxScanTime, "scan-time", 0, "Max number of seconds to perform scan. 0 means infinite")
	DiscoverCmd.Flags().BoolVar(&discoverVerbose, "verbose", false, "Verbose output during scan")
}

func doBTScan(logger log.Logger) error {

	if discoverMaxScanTime == 0 {
		discoverMaxScanTime = math.MaxUint32
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	adapter := bluetooth.DefaultAdapter

	if err := adapter.Enable(); err != nil {
		return err
	}

	scanTimer := time.NewTimer(time.Second * time.Duration(discoverMaxScanTime))
	go func() {
		select {
		case <-sigs:
			scanTimer.Stop()
		case <-scanTimer.C:
		}
		adapter.StopScan()
	}()

	scanResults := make(map[string]bluetooth.ScanResult)

	if discoverMaxScanTime == math.MaxUint32 {
		level.Info(logger).Log("msg", "Scanning forever [CTRL+C to stop]")
	} else {
		level.Info(logger).Log("msg", fmt.Sprintf("Scanning for %d second(s)", discoverMaxScanTime))
	}

	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		addr := result.Address.String()
		if _, prs := scanResults[addr]; !prs {
			level.Info(logger).Log("msg", "Found device", "address", addr, "rssi", result.RSSI, "name", result.LocalName())
			scanResults[addr] = result
		}
	})
	if err != nil {
		return err
	}

	return nil
}

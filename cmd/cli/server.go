package main

import (
	"fmt"
	"os"

	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/server"
	"github.com/spf13/cobra"
)

var (
	serverPort    int
	serverTarget  string
	serverVerbose bool
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start an HTTP server to control the iDot display",
	Long: `Start an HTTP server that exposes:
  - API endpoints at /api/* for programmatic control
  - Web console at / for browser-based control

Examples:
  idm-cli server
  idm-cli server --port 8080
  idm-cli server --target AA:BB:CC:DD:EE:FF --port 3010`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(serverVerbose)
		if err := runServer(logger); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	ServerCmd.Flags().IntVar(&serverPort, "port", 3010, "HTTP server port")
	ServerCmd.Flags().StringVar(&serverTarget, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")
	ServerCmd.Flags().BoolVar(&serverVerbose, "verbose", false, "Enable verbose debug logging")
}

func runServer(logger interface{ Log(...interface{}) error }) error {
	srv := server.New(logger, serverTarget, serverPort)
	return srv.Start()
}

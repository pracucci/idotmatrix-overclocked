// Package server provides an HTTP server for controlling the iDotMatrix display.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
)

// Server is the HTTP server for controlling the iDotMatrix display.
type Server struct {
	logger     log.Logger
	targetAddr string
	port       int

	device       *protocol.Device
	mu           sync.Mutex // Serializes device access
	connected    bool
	reconnecting bool
	snakeManager *SnakeManager
}

// New creates a new Server instance.
func New(logger log.Logger, targetAddr string, port int) *Server {
	return &Server{
		logger:       logger,
		targetAddr:   targetAddr,
		port:         port,
		snakeManager: NewSnakeManager(),
	}
}

// Start starts the HTTP server and connects to the device in the background.
func (s *Server) Start() error {
	// Start device connection in the background
	go s.connectInBackground()

	// Set up HTTP routes
	mux := http.NewServeMux()
	s.registerAPIRoutes(mux)
	s.registerConsoleRoutes(mux)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", s.port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Handle graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		level.Info(s.logger).Log("msg", "Shutting down server")

		// Stop any active snake game
		s.snakeManager.StopSession()

		// Shutdown HTTP server
		if err := httpServer.Shutdown(context.Background()); err != nil {
			level.Error(s.logger).Log("msg", "HTTP server shutdown error", "err", err)
		}

		// Disconnect device
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.device != nil {
			if err := s.device.Disconnect(); err != nil {
				level.Error(s.logger).Log("msg", "Failed to disconnect device", "err", err)
			} else {
				level.Info(s.logger).Log("msg", "Disconnected from device")
			}
		}
	}()

	level.Info(s.logger).Log("msg", "Starting HTTP server", "addr", addr)
	fmt.Printf("iDotMatrix HTTP Server running at http://localhost:%d\n", s.port)
	fmt.Printf("  Web Console: http://localhost:%d/\n", s.port)
	fmt.Printf("  API Status:  http://localhost:%d/api/status\n", s.port)

	// Open web console in browser
	url := fmt.Sprintf("http://localhost:%d/", s.port)
	_ = exec.Command("open", url).Start()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// connectInBackground attempts to connect to the device in a background loop.
// It keeps retrying until a connection is established.
func (s *Server) connectInBackground() {
	s.mu.Lock()
	if s.reconnecting {
		s.mu.Unlock()
		return
	}
	s.reconnecting = true
	s.mu.Unlock()

	level.Info(s.logger).Log("msg", "Starting device connection in background", "target", s.targetAddr)

	for {
		s.mu.Lock()
		err := s.connect()
		if err == nil {
			s.reconnecting = false
			s.mu.Unlock()
			level.Info(s.logger).Log("msg", "Device connected")
			return
		}
		s.mu.Unlock()

		level.Warn(s.logger).Log("msg", "Device connection failed, retrying", "err", err)
		time.Sleep(2 * time.Second)
	}
}

// connect establishes a connection to the device.
// Must be called with s.mu held.
func (s *Server) connect() error {
	level.Debug(s.logger).Log("msg", "Attempting to connect to device", "target", s.targetAddr)
	s.device = protocol.NewDevice(s.logger)
	if err := s.device.Connect(s.targetAddr); err != nil {
		return err
	}
	s.connected = true
	return nil
}

// reconnect attempts to reconnect to the device in a loop until successful.
func (s *Server) reconnect() {
	s.mu.Lock()
	if s.reconnecting {
		s.mu.Unlock()
		return
	}
	s.reconnecting = true
	s.connected = false

	// Try to disconnect cleanly first
	if s.device != nil {
		_ = s.device.Disconnect()
	}
	s.mu.Unlock()

	level.Info(s.logger).Log("msg", "Starting reconnection loop")

	for {
		time.Sleep(2 * time.Second)

		s.mu.Lock()
		err := s.connect()
		if err == nil {
			s.reconnecting = false
			s.mu.Unlock()
			level.Info(s.logger).Log("msg", "Reconnected to device")
			return
		}

		level.Warn(s.logger).Log("msg", "Reconnection failed, retrying", "err", err)
		s.mu.Unlock()
	}
}

// isConnectionError checks if an error indicates the device connection was lost.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "disconnected") ||
		strings.Contains(errStr, "not connected") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout")
}

// withDevice executes a function with exclusive device access.
// If a connection error is detected, it triggers automatic reconnection.
func (s *Server) withDevice(fn func(*protocol.Device) error) error {
	s.mu.Lock()
	if !s.connected {
		s.mu.Unlock()
		return fmt.Errorf("device not connected, waiting for connection...")
	}
	device := s.device
	s.mu.Unlock()

	err := fn(device)
	if err != nil && isConnectionError(err) {
		s.mu.Lock()
		s.connected = false
		s.mu.Unlock()
		go s.reconnect()
		return fmt.Errorf("device disconnected, reconnecting...")
	}
	return err
}

// IsConnected returns whether the server is connected to the device.
func (s *Server) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connected
}

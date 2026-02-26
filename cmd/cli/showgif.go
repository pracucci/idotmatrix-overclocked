package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
)

const (
	showgifDisplaySize    = 64
	showgifMaxFrames      = 64
	showgifMaxDurationMs  = 2000
	showgifMinFrameTimeMs = 16 // ~60 FPS limit
)

var showgifTargetAddr string
var showgifGifFile string
var showgifVerbose bool
var showgifMirrored bool
var showgifBrightness int

var ShowgifCmd = &cobra.Command{
	Use:   "showgif",
	Short: "Shows an animated GIF on the 64x64 iDot display",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(showgifVerbose)
		if err := doShowGIF(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	ShowgifCmd.Flags().StringVar(&showgifTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")

	ShowgifCmd.Flags().StringVar(&showgifGifFile, "gif-file", "", "Path to a 64x64 animated GIF file")
	ShowgifCmd.MarkFlagRequired("gif-file")

	ShowgifCmd.Flags().BoolVar(&showgifVerbose, "verbose", false, "Enable verbose debug logging")
	ShowgifCmd.Flags().BoolVar(&showgifMirrored, "mirrored", false, "Mirror the image horizontally")
	ShowgifCmd.Flags().IntVar(&showgifBrightness, "brightness", 100, "Brightness level (0-100)")
}

// loadAndReencodeGIF loads a GIF, re-composites frames, and re-encodes it for the device.
func loadAndReencodeGIF(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	g, err := gif.DecodeAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode GIF: %w", err)
	}

	// Validate dimensions
	if g.Config.Width != showgifDisplaySize || g.Config.Height != showgifDisplaySize {
		return nil, fmt.Errorf("GIF is %dx%d, expected %dx%d", g.Config.Width, g.Config.Height, showgifDisplaySize, showgifDisplaySize)
	}

	numFrames := len(g.Image)
	if numFrames == 0 {
		return nil, fmt.Errorf("GIF has no frames")
	}
	if numFrames > showgifMaxFrames {
		fmt.Printf("Warning: GIF has %d frames, limiting to %d\n", numFrames, showgifMaxFrames)
		numFrames = showgifMaxFrames
	}

	// Calculate and adjust frame delays
	totalDurationMs := 0
	delays := make([]int, numFrames)
	for i := 0; i < numFrames; i++ {
		delay := g.Delay[i]
		if delay < showgifMinFrameTimeMs/10 {
			delay = showgifMinFrameTimeMs / 10 // Minimum 16ms (delay is in 1/100s)
		}
		delays[i] = delay
		totalDurationMs += delay * 10
	}

	if totalDurationMs > showgifMaxDurationMs {
		fmt.Printf("Warning: GIF duration %dms exceeds %dms limit\n", totalDurationMs, showgifMaxDurationMs)
	}

	// Re-composite and re-encode frames
	canvas := image.NewRGBA(image.Rect(0, 0, showgifDisplaySize, showgifDisplaySize))
	newFrames := make([]*image.Paletted, numFrames)

	for i := 0; i < numFrames; i++ {
		frame := g.Image[i]
		bounds := frame.Bounds()

		// Composite frame onto canvas
		draw.Draw(canvas, bounds, frame, bounds.Min, draw.Over)

		// Create new paletted frame from canvas
		palettedFrame := image.NewPaletted(image.Rect(0, 0, showgifDisplaySize, showgifDisplaySize), palette.Plan9)
		draw.Draw(palettedFrame, palettedFrame.Bounds(), canvas, image.Point{}, draw.Src)
		newFrames[i] = palettedFrame

		// Handle disposal (disposal=2 means restore to background)
		if i < len(g.Disposal) && g.Disposal[i] == gif.DisposalBackground {
			draw.Draw(canvas, bounds, image.Black, image.Point{}, draw.Src)
		}
	}

	// Re-encode GIF with loop forever and disposal=2 (restore to background)
	newGIF := &gif.GIF{
		Image:     newFrames,
		Delay:     delays,
		LoopCount: 0, // loop forever
		Disposal:  make([]byte, numFrames),
	}
	// Set disposal=2 (DisposalBackground) for all frames
	for i := range newGIF.Disposal {
		newGIF.Disposal[i] = gif.DisposalBackground
	}

	// Encode to bytes
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, newGIF); err != nil {
		return nil, fmt.Errorf("failed to re-encode GIF: %w", err)
	}

	fmt.Printf("Loaded GIF: %d frames, %dms total, re-encoded to %d bytes\n", numFrames, totalDurationMs, buf.Len())

	return buf.Bytes(), nil
}

func doShowGIF(logger log.Logger) error {
	if len(showgifGifFile) == 0 {
		return fmt.Errorf("missing --gif-file option")
	}

	gifData, err := loadAndReencodeGIF(showgifGifFile)
	if err != nil {
		return err
	}

	g, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err != nil {
		return fmt.Errorf("failed to decode GIF: %w", err)
	}
	if showgifMirrored {
		g = graphic.MirrorGIFHorizontal(g)
	}
	g = graphic.AdjustBrightnessGIF(g, showgifBrightness)
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return fmt.Errorf("failed to re-encode GIF: %w", err)
	}
	gifData = buf.Bytes()

	device := protocol.NewDevice(logger)
	if err = device.Connect(showgifTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	// Send GIF directly - no SetDrawMode needed
	if err := protocol.SendGIF(device, gifData, logger); err != nil {
		return err
	}

	// Allow time for final writes to complete
	time.Sleep(500 * time.Millisecond)

	return nil
}

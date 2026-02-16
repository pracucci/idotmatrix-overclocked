package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"math/rand"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
)

const (
	fireDisplaySize = 64
	fireNumFrames   = 32
	fireFrameDelay  = 5 // 50ms per frame (delay is in 1/100s)
	firePaletteSize = 37
)

var fireTargetAddr string
var fireVerbose bool

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
}

// DOOM fire palette: black → red → orange → yellow → white
var firePalette = []color.RGBA{
	{0x07, 0x07, 0x07, 0xFF},
	{0x1F, 0x07, 0x07, 0xFF},
	{0x2F, 0x0F, 0x07, 0xFF},
	{0x47, 0x0F, 0x07, 0xFF},
	{0x57, 0x17, 0x07, 0xFF},
	{0x67, 0x1F, 0x07, 0xFF},
	{0x77, 0x1F, 0x07, 0xFF},
	{0x8F, 0x27, 0x07, 0xFF},
	{0x9F, 0x2F, 0x07, 0xFF},
	{0xAF, 0x3F, 0x07, 0xFF},
	{0xBF, 0x47, 0x07, 0xFF},
	{0xC7, 0x47, 0x07, 0xFF},
	{0xDF, 0x4F, 0x07, 0xFF},
	{0xDF, 0x57, 0x07, 0xFF},
	{0xDF, 0x57, 0x07, 0xFF},
	{0xD7, 0x5F, 0x07, 0xFF},
	{0xD7, 0x5F, 0x07, 0xFF},
	{0xD7, 0x67, 0x0F, 0xFF},
	{0xCF, 0x6F, 0x0F, 0xFF},
	{0xCF, 0x77, 0x0F, 0xFF},
	{0xCF, 0x7F, 0x0F, 0xFF},
	{0xCF, 0x87, 0x17, 0xFF},
	{0xC7, 0x87, 0x17, 0xFF},
	{0xC7, 0x8F, 0x17, 0xFF},
	{0xC7, 0x97, 0x1F, 0xFF},
	{0xBF, 0x9F, 0x1F, 0xFF},
	{0xBF, 0x9F, 0x1F, 0xFF},
	{0xBF, 0xA7, 0x27, 0xFF},
	{0xBF, 0xA7, 0x27, 0xFF},
	{0xBF, 0xAF, 0x2F, 0xFF},
	{0xB7, 0xAF, 0x2F, 0xFF},
	{0xB7, 0xB7, 0x2F, 0xFF},
	{0xB7, 0xB7, 0x37, 0xFF},
	{0xCF, 0xCF, 0x6F, 0xFF},
	{0xDF, 0xDF, 0x9F, 0xFF},
	{0xEF, 0xEF, 0xC7, 0xFF},
	{0xF0, 0xE0, 0x50, 0xFF}, // warm yellow (heat source)
}

func generateFireGIF() []byte {
	rand.Seed(time.Now().UnixNano())

	// Initialize fire buffer
	firePixels := make([]int, fireDisplaySize*fireDisplaySize)

	// Set bottom row to max heat (white)
	for x := 0; x < fireDisplaySize; x++ {
		firePixels[(fireDisplaySize-1)*fireDisplaySize+x] = firePaletteSize - 1
	}

	// Warmup: run simulation until fire reaches steady state
	// This ensures the GIF starts with flames already burning
	for i := 0; i < 200; i++ {
		spreadFireFrame(firePixels)
	}

	// Build color palette for GIF
	gifPalette := make(color.Palette, firePaletteSize)
	for i, c := range firePalette {
		gifPalette[i] = c
	}

	frames := make([]*image.Paletted, fireNumFrames)
	delays := make([]int, fireNumFrames)

	for t := 0; t < fireNumFrames; t++ {
		// Simulate fire spread multiple times per frame
		for i := 0; i < 4; i++ {
			spreadFireFrame(firePixels)
		}

		// Create frame from buffer
		frame := image.NewPaletted(image.Rect(0, 0, fireDisplaySize, fireDisplaySize), gifPalette)
		for y := 0; y < fireDisplaySize; y++ {
			for x := 0; x < fireDisplaySize; x++ {
				idx := firePixels[y*fireDisplaySize+x]
				frame.SetColorIndex(x, y, uint8(idx))
			}
		}
		frames[t] = frame
		delays[t] = fireFrameDelay
	}

	g := &gif.GIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: 0,
	}
	var buf bytes.Buffer
	gif.EncodeAll(&buf, g)
	return buf.Bytes()
}

func spreadFireFrame(firePixels []int) {
	for x := 0; x < fireDisplaySize; x++ {
		for y := 1; y < fireDisplaySize; y++ {
			spreadFire(firePixels, y*fireDisplaySize+x)
		}
	}
}

func spreadFire(firePixels []int, src int) {
	pixel := firePixels[src]
	if pixel == 0 {
		firePixels[src-fireDisplaySize] = 0
		return
	}
	randVal := rand.Intn(8)
	dst := src - randVal + 1
	if dst < fireDisplaySize {
		dst = fireDisplaySize
	}
	// Decay rate: average 1.0 per row so flames die out in ~36 pixels
	// Values: 0,0,1,1,1,1,2,2 -> average 8/8 = 1.0
	decay := (randVal + 2) / 4
	newVal := pixel - decay
	if newVal < 0 {
		newVal = 0
	}
	firePixels[dst-fireDisplaySize] = newVal
}

func doFire(logger log.Logger) error {
	fmt.Println("Generating DOOM fire animation...")
	gifData := generateFireGIF()
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

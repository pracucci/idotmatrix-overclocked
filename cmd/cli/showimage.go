package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pracucci/idotmatrix-overclocked/pkg/logging"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/spf13/cobra"
	xdraw "golang.org/x/image/draw"
)

var showimageTargetAddr string
var showimageImageFile string
var showimageDisplaySize int
var showimageVerbose bool

var ShowimageCmd = &cobra.Command{
	Use:   "showimage",
	Short: "Shows the supplied image file on the iDot display",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.NewLogger(showimageVerbose)
		if err := doShowImage(logger); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func init() {
	ShowimageCmd.Flags().StringVar(&showimageTargetAddr, "target", "", "Target iDot display MAC address (auto-discovers if not specified)")

	ShowimageCmd.Flags().StringVar(&showimageImageFile, "image-file", "", "Path to an image file (PNG, JPEG, GIF)")
	ShowimageCmd.MarkFlagRequired("image-file")

	ShowimageCmd.Flags().IntVar(&showimageDisplaySize, "size", 64, "Display size (32 or 64)")
	ShowimageCmd.Flags().BoolVar(&showimageVerbose, "verbose", false, "Enable verbose debug logging")
}

// loadAndConvertImage loads an image file and converts it to raw RGB data
func loadAndConvertImage(filePath string, expectedSize int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != expectedSize || bounds.Dy() != expectedSize {
		fmt.Printf("Resizing image from %dx%d to %dx%d\n", bounds.Dx(), bounds.Dy(), expectedSize, expectedSize)
		dst := image.NewRGBA(image.Rect(0, 0, expectedSize, expectedSize))
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, xdraw.Over, nil)
		img = dst
		bounds = img.Bounds()
	}

	rgbData := make([]byte, expectedSize*expectedSize*3)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rgbData[idx] = uint8(r >> 8)
			rgbData[idx+1] = uint8(g >> 8)
			rgbData[idx+2] = uint8(b >> 8)
			idx += 3
		}
	}

	return rgbData, nil
}

func doShowImage(logger log.Logger) error {
	if len(showimageImageFile) == 0 {
		return fmt.Errorf("missing --image-file option")
	}
	if showimageDisplaySize != 32 && showimageDisplaySize != 64 {
		return fmt.Errorf("invalid display size: %d (must be 32 or 64)", showimageDisplaySize)
	}

	rgbData, err := loadAndConvertImage(showimageImageFile, showimageDisplaySize)
	if err != nil {
		return err
	}

	device := protocol.NewDevice(logger)
	if err = device.Connect(showimageTargetAddr); err != nil {
		return err
	}
	defer func() {
		if err := device.Disconnect(); err != nil {
			level.Error(logger).Log("msg", "Failed to disconnect", "err", err)
		}
	}()

	if err := protocol.SetDrawMode(device, 1); err != nil {
		return err
	}

	if err := protocol.SendImage(device, rgbData); err != nil {
		return err
	}

	// Allow time for BLE writes to complete before disconnecting
	time.Sleep(500 * time.Millisecond)

	return nil
}

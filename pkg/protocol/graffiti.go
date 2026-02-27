package protocol

import (
	"time"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// PacketDelay is the minimum delay between consecutive BLE packets when using
// SetPixels for real-time rendering. This prevents overwhelming the device's
// BLE receive buffer and ensures reliable pixel updates.
const PacketDelay = 50 * time.Millisecond

// MaxPixelsPerPacket is the maximum number of coordinate pairs in one SetPixels packet
const MaxPixelsPerPacket = 255

// SetPixel sends a single pixel to the display using the graffiti protocol.
// Coordinates are 0-63 for a 64x64 display.
func SetPixel(d DeviceConnection, x, y int, r, g, b uint8) error {
	payload := []byte{
		0x0A, 0x00, 0x05, 0x01, 0x00, // header
		r, g, b, // RGB
		byte(x), byte(y), // coordinates
	}
	return d.WritePacket(payload)
}

// SetPixels sends multiple pixels with the same color to the display.
// Uses the multi-pixel protocol: [size_lsb, size_msb, 0x05, 0x01, 0x00, R, G, B, X1, Y1, X2, Y2, ...]
// Maximum 255 coordinate pairs per packet.
func SetPixels(d DeviceConnection, color graphic.Color, points []graphic.Point) error {
	if len(points) == 0 {
		return nil
	}
	if len(points) > 255 {
		points = points[:255]
	}

	// Size = 8 (header + RGB) + 2 * num_pixels (coordinates)
	size := 8 + 2*len(points)
	sizeLSB := byte(size & 0xFF)
	sizeMSB := byte((size >> 8) & 0xFF)

	payload := make([]byte, size)
	payload[0] = sizeLSB
	payload[1] = sizeMSB
	payload[2] = 0x05 // graffiti command
	payload[3] = 0x01
	payload[4] = 0x00
	payload[5] = color[0]
	payload[6] = color[1]
	payload[7] = color[2]

	// Add coordinates
	offset := 8
	for _, p := range points {
		payload[offset] = byte(p.X)
		payload[offset+1] = byte(p.Y)
		offset += 2
	}

	return d.WritePacket(payload)
}

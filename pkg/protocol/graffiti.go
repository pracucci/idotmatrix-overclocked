package protocol

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

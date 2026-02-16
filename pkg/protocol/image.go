package protocol

import (
	"bytes"
	"encoding/binary"
)

// SetDrawMode sends set draw mode to display.
func SetDrawMode(d DeviceConnection, mode int) error {
	return WriteData(d, []byte{5, 0, 4, 1, uint8(mode)})
}

// SendImage sends an image to the display. Only makes sense after a call to SetDrawMode(1).
// imageData should be raw RGB data (3 bytes per pixel).
func SendImage(d DeviceConnection, imageData []byte) error {
	const headerSize = 9

	// Split image data into 4096-byte chunks
	chunks := chunkBuffer(imageData, 4096)

	for ci, ch := range chunks {
		packet := new(bytes.Buffer)

		// Header (9 bytes):
		// Bytes 0-1: Packet length (chunk size + header size)
		binary.Write(packet, binary.LittleEndian, uint16(len(ch)+headerSize))
		// Bytes 2-3: Command/Subcommand (both 0x00)
		binary.Write(packet, binary.LittleEndian, uint8(0))
		binary.Write(packet, binary.LittleEndian, uint8(0))
		// Byte 4: Packet type (0x00 for first, 0x02 for continuation)
		if ci > 0 {
			binary.Write(packet, binary.LittleEndian, uint8(2))
		} else {
			binary.Write(packet, binary.LittleEndian, uint8(0))
		}
		// Bytes 5-8: Total image data length
		binary.Write(packet, binary.LittleEndian, int32(len(imageData)))
		// Chunk data
		binary.Write(packet, binary.LittleEndian, ch)

		if err := WriteData(d, packet.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

// chunkBuffer chunks the supplied data buffer to chunkSize slices.
func chunkBuffer(data []byte, chunkSize int) [][]byte {
	chunks := make([][]byte, 0)

	cursor := 0
	remaining := len(data)
	for remaining > 0 {
		wl := min(chunkSize, remaining)
		chunks = append(chunks, data[cursor:cursor+wl])
		cursor += wl
		remaining -= wl
	}
	return chunks
}

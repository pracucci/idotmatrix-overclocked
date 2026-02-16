package protocol

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	gifHeaderSize          = 16
	gifChunkSize           = 4096
	gifBLEPacketSize       = 509 // BLE packet size for GIF uploads
	gifTypeNoTimeSignature = 12  // NO_TIME_SIGNATURE GIF type
)

// SendGIF sends an animated GIF to the display.
// gifData should be the raw GIF file bytes (re-encoded GIF).
func SendGIF(d DeviceConnection, gifData []byte, logger log.Logger) error {
	// Drain any stale notifications from previous operations
	d.DrainResponses()

	// Brief stabilization delay after connection
	time.Sleep(100 * time.Millisecond)

	// Calculate CRC32 of entire GIF data (same value used in all chunk headers)
	crc := crc32.ChecksumIEEE(gifData)

	level.Debug(logger).Log("msg", "GIF upload starting", "size", len(gifData), "crc32", fmt.Sprintf("0x%08X", crc))

	// Split GIF data into 4096-byte chunks
	chunks := chunkBuffer(gifData, gifChunkSize)

	level.Debug(logger).Log("msg", "Split into chunks", "chunks", len(chunks))

	for ci, chunk := range chunks {
		// Build 16-byte header for this chunk
		header := make([]byte, gifHeaderSize)

		// Bytes 0-1: Packet length (chunk + header size) - little-endian
		packetLen := uint16(len(chunk) + gifHeaderSize)
		binary.LittleEndian.PutUint16(header[0:2], packetLen)

		// Byte 2: Command type (1 for GIF)
		header[2] = 1

		// Byte 3: Sub-command (0)
		header[3] = 0

		// Byte 4: Packet sequence indicator
		// 0x00 = first packet
		// 0x02 = continuation packet
		if ci > 0 {
			header[4] = 2 // Continuation
		} else {
			header[4] = 0 // First packet
		}

		// Bytes 5-8: Total GIF data length - little-endian
		binary.LittleEndian.PutUint32(header[5:9], uint32(len(gifData)))

		// Bytes 9-12: CRC32 - little-endian
		binary.LittleEndian.PutUint32(header[9:13], crc)

		// Bytes 13-14: Time signature = 0 (for gif_type=12)
		header[13] = 0
		header[14] = 0

		// Byte 15: GIF type = 12 (NO_TIME_SIGNATURE)
		header[15] = gifTypeNoTimeSignature

		level.Debug(logger).Log("msg", "Sending chunk", "chunk", ci+1, "total", len(chunks), "header", fmt.Sprintf("%v", header))

		// Combine header + chunk data into large packet
		largePacket := append(header, chunk...)

		// Split into BLE packets
		blePackets := chunkBuffer(largePacket, gifBLEPacketSize)

		level.Debug(logger).Log("msg", "BLE packets to send", "chunk", ci+1, "packets", len(blePackets))

		// Send all BLE packets for this chunk
		for pi, pkt := range blePackets {
			level.Debug(logger).Log("msg", "Sending BLE packet", "packet", pi+1, "total", len(blePackets), "bytes", len(pkt))
			if err := d.WritePacket(pkt); err != nil {
				return fmt.Errorf("failed to send BLE packet %d: %w", pi+1, err)
			}
			// Small delay between packets to let device process
			time.Sleep(10 * time.Millisecond)
		}

		// Read response after sending all packets for this chunk
		level.Debug(logger).Log("msg", "Waiting for response", "chunk", ci+1)
		response, err := d.ReadResponse()
		if err != nil {
			return fmt.Errorf("chunk %d: read response failed: %w", ci+1, err)
		}
		level.Debug(logger).Log("msg", "Response received", "response", fmt.Sprintf("%v", response), "hex", fmt.Sprintf("%X", response))

		// Expected responses:
		// [5 0 1 0 1] = chunk OK, continue
		// [5 0 1 0 3] = upload complete
		isFirstChunk := ci == 0
		isLastChunk := ci == len(chunks)-1
		if len(response) >= 5 && response[0] == 5 && response[1] == 0 && response[2] == 1 && response[3] == 0 {
			if response[4] == 1 {
				level.Debug(logger).Log("msg", "Chunk received, continuing")
			} else if response[4] == 3 {
				if isLastChunk {
					level.Debug(logger).Log("msg", "Upload complete")
				} else if isFirstChunk {
					// Device already has this GIF cached (recognized by CRC32)
					level.Debug(logger).Log("msg", "GIF already cached on device, upload complete")
					return nil
				} else {
					return fmt.Errorf("chunk %d/%d: device returned 'upload complete' prematurely (expected 'continue')", ci+1, len(chunks))
				}
			} else {
				return fmt.Errorf("chunk %d: unexpected response code: %d", ci+1, response[4])
			}
		} else {
			return fmt.Errorf("chunk %d: unexpected response format: %v", ci+1, response)
		}

	}

	return nil
}

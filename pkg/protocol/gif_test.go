package protocol

import (
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendGIF(t *testing.T) {
	// Response codes
	responseOKContinue := []byte{5, 0, 1, 0, 1} // chunk OK, continue
	responseComplete := []byte{5, 0, 1, 0, 3}   // upload complete

	t.Run("single chunk completes upload", func(t *testing.T) {
		mock := &DeviceConnectionMock{}
		mock.AddResponse(responseComplete)

		gifData := make([]byte, 100) // Small GIF, fits in one chunk
		for i := range gifData {
			gifData[i] = byte(i % 256)
		}

		err := SendGIF(mock, gifData, log.NewNopLogger())
		require.NoError(t, err)
		assert.True(t, mock.DrainCalled, "DrainResponses should be called")
		require.NotEmpty(t, mock.WrittenPackets, "expected BLE packets to be sent")

		// Verify first packet contains correct header
		firstPacket := mock.WrittenPackets[0]
		require.GreaterOrEqual(t, len(firstPacket), gifHeaderSize, "first packet too short")

		// Verify CRC32 in header
		expectedCRC := crc32.ChecksumIEEE(gifData)
		headerCRC := binary.LittleEndian.Uint32(firstPacket[9:13])
		assert.Equal(t, expectedCRC, headerCRC, "CRC32 mismatch")

		// Verify total length in header
		totalLen := binary.LittleEndian.Uint32(firstPacket[5:9])
		assert.Equal(t, uint32(len(gifData)), totalLen, "total length mismatch")
	})

	t.Run("multiple chunks with continue responses", func(t *testing.T) {
		mock := &DeviceConnectionMock{}
		mock.AddResponse(responseOKContinue) // First chunk: continue
		mock.AddResponse(responseComplete)   // Second chunk: complete

		gifData := make([]byte, 5000) // Larger than 4096, needs 2 chunks
		for i := range gifData {
			gifData[i] = byte(i % 256)
		}

		err := SendGIF(mock, gifData, log.NewNopLogger())
		require.NoError(t, err)

		// Count chunks by looking for packet sequence indicators
		// First chunk packets have header[4] = 0, continuation have header[4] = 2
		firstChunkPackets := 0
		continuationPackets := 0
		for _, pkt := range mock.WrittenPackets {
			if len(pkt) >= gifHeaderSize {
				if pkt[4] == 0 {
					firstChunkPackets++
				} else if pkt[4] == 2 {
					continuationPackets++
				}
			}
		}

		assert.Greater(t, firstChunkPackets, 0, "expected at least one first chunk packet")
		assert.Greater(t, continuationPackets, 0, "expected at least one continuation chunk packet")
	})

	t.Run("cached on device stops after first chunk", func(t *testing.T) {
		mock := &DeviceConnectionMock{}
		mock.AddResponse(responseComplete) // Device says complete on first chunk (cached)

		gifData := make([]byte, 5000) // Large enough for 2 chunks
		for i := range gifData {
			gifData[i] = byte(i % 256)
		}

		err := SendGIF(mock, gifData, log.NewNopLogger())
		require.NoError(t, err)

		// Should only have packets for first chunk since device said complete
		// Check that no continuation packets (header[4] = 2) were sent
		for _, pkt := range mock.WrittenPackets {
			if len(pkt) >= gifHeaderSize {
				assert.NotEqual(t, byte(2), pkt[4], "sent continuation chunk but device indicated cached")
			}
		}
	})
}

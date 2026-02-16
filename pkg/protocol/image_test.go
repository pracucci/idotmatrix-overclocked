package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDrawMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     int
		expected []byte
	}{
		{
			name:     "mode 0",
			mode:     0,
			expected: []byte{5, 0, 4, 1, 0},
		},
		{
			name:     "mode 1",
			mode:     1,
			expected: []byte{5, 0, 4, 1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeviceConnectionMock{}
			err := SetDrawMode(mock, tt.mode)
			require.NoError(t, err)
			require.Len(t, mock.WrittenPackets, 1)
			assert.Equal(t, tt.expected, mock.WrittenPackets[0])
		})
	}
}

func TestSendImage(t *testing.T) {
	t.Run("single chunk for small image", func(t *testing.T) {
		mock := &DeviceConnectionMock{}
		imageData := make([]byte, 100) // Small image
		for i := range imageData {
			imageData[i] = byte(i % 256)
		}

		err := SendImage(mock, imageData)
		require.NoError(t, err)
		require.Len(t, mock.WrittenPackets, 1)

		packet := mock.WrittenPackets[0]
		// Verify header
		packetLen := binary.LittleEndian.Uint16(packet[0:2])
		assert.Equal(t, uint16(100+9), packetLen)
		assert.Equal(t, byte(0), packet[4], "packet type should be 0 (first packet)")

		totalLen := binary.LittleEndian.Uint32(packet[5:9])
		assert.Equal(t, uint32(100), totalLen)
	})

	t.Run("multiple chunks for large image", func(t *testing.T) {
		mock := &DeviceConnectionMock{}
		imageData := make([]byte, 5000) // Larger than 4096
		for i := range imageData {
			imageData[i] = byte(i % 256)
		}

		err := SendImage(mock, imageData)
		require.NoError(t, err)

		// Reconstruct the full data sent by concatenating all BLE packets
		var allData bytes.Buffer
		for _, pkt := range mock.WrittenPackets {
			allData.Write(pkt)
		}

		// First protocol chunk: header + 4096 bytes of image data
		firstChunkSize := 9 + 4096
		// Second protocol chunk: header + remaining 904 bytes
		secondChunkSize := 9 + 904
		expectedTotalSize := firstChunkSize + secondChunkSize
		assert.Equal(t, expectedTotalSize, allData.Len(), "total bytes sent")

		data := allData.Bytes()

		// Verify first chunk header
		assert.Equal(t, byte(0), data[4], "first chunk type should be 0")
		totalLen1 := binary.LittleEndian.Uint32(data[5:9])
		assert.Equal(t, uint32(5000), totalLen1)

		// Verify second chunk header (starts after first chunk)
		secondChunkStart := firstChunkSize
		assert.Equal(t, byte(2), data[secondChunkStart+4], "second chunk type should be 2 (continuation)")
		totalLen2 := binary.LittleEndian.Uint32(data[secondChunkStart+5 : secondChunkStart+9])
		assert.Equal(t, uint32(5000), totalLen2)
	})
}

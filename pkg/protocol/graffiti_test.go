package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetPixel(t *testing.T) {
	tests := []struct {
		name     string
		x, y     int
		r, g, b  uint8
		expected []byte
	}{
		{
			name:     "pixel at origin with red",
			x:        0, y: 0,
			r:        255, g: 0, b: 0,
			expected: []byte{0x0A, 0x00, 0x05, 0x01, 0x00, 255, 0, 0, 0, 0},
		},
		{
			name:     "pixel at corner with green",
			x:        63, y: 63,
			r:        0, g: 255, b: 0,
			expected: []byte{0x0A, 0x00, 0x05, 0x01, 0x00, 0, 255, 0, 63, 63},
		},
		{
			name:     "pixel at center with blue",
			x:        32, y: 32,
			r:        0, g: 0, b: 255,
			expected: []byte{0x0A, 0x00, 0x05, 0x01, 0x00, 0, 0, 255, 32, 32},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeviceConnectionMock{}
			err := SetPixel(mock, tt.x, tt.y, tt.r, tt.g, tt.b)
			require.NoError(t, err)
			require.Len(t, mock.WrittenPackets, 1)
			assert.Equal(t, tt.expected, mock.WrittenPackets[0])
		})
	}
}

package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
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

func TestSetPixels(t *testing.T) {
	tests := []struct {
		name     string
		color    graphic.Color
		points   []graphic.Point
		expected []byte
	}{
		{
			name:     "empty points returns no packet",
			color:    graphic.Color{255, 0, 0},
			points:   []graphic.Point{},
			expected: nil,
		},
		{
			name:   "single pixel",
			color:  graphic.Color{255, 0, 0},
			points: []graphic.Point{{X: 10, Y: 20}},
			// size = 8 + 2*1 = 10, sizeLSB=10, sizeMSB=0
			expected: []byte{0x0A, 0x00, 0x05, 0x01, 0x00, 255, 0, 0, 10, 20},
		},
		{
			name:   "multiple pixels same color",
			color:  graphic.Color{0, 255, 0},
			points: []graphic.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}},
			// size = 8 + 2*3 = 14, sizeLSB=14, sizeMSB=0
			expected: []byte{0x0E, 0x00, 0x05, 0x01, 0x00, 0, 255, 0, 0, 0, 1, 1, 2, 2},
		},
		{
			name:   "blue color",
			color:  graphic.Color{0, 0, 255},
			points: []graphic.Point{{X: 63, Y: 63}},
			expected: []byte{0x0A, 0x00, 0x05, 0x01, 0x00, 0, 0, 255, 63, 63},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeviceConnectionMock{}
			err := SetPixels(mock, tt.color, tt.points)
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Len(t, mock.WrittenPackets, 0)
			} else {
				require.Len(t, mock.WrittenPackets, 1)
				assert.Equal(t, tt.expected, mock.WrittenPackets[0])
			}
		})
	}
}

func TestSetPixelsTruncatesAt255(t *testing.T) {
	// Create 300 points - should be truncated to 255
	points := make([]graphic.Point, 300)
	for i := 0; i < 300; i++ {
		points[i] = graphic.Point{X: i % 64, Y: i / 64}
	}

	mock := &DeviceConnectionMock{}
	err := SetPixels(mock, graphic.Color{255, 255, 255}, points)
	require.NoError(t, err)
	require.Len(t, mock.WrittenPackets, 1)

	// size = 8 + 2*255 = 518 = 0x206
	// sizeLSB = 0x06, sizeMSB = 0x02
	packet := mock.WrittenPackets[0]
	assert.Equal(t, byte(0x06), packet[0], "sizeLSB should be 0x06")
	assert.Equal(t, byte(0x02), packet[1], "sizeMSB should be 0x02")

	// Packet should have 8 header bytes + 255*2 coordinate bytes = 518 bytes
	assert.Len(t, packet, 518)
}

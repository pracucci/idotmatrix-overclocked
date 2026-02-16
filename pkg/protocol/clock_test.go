package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

func TestSetClockMode(t *testing.T) {
	tests := []struct {
		name        string
		style       int
		visibleDate bool
		hour24      bool
		color       graphic.Color
		expected    []byte
	}{
		{
			name:        "default style without date or 24h",
			style:       ClockDefault,
			visibleDate: false,
			hour24:      false,
			color:       graphic.Color{255, 255, 255},
			expected:    []byte{8, 0, 6, 1, 0, 255, 255, 255},
		},
		{
			name:        "with visible date",
			style:       ClockDefault,
			visibleDate: true,
			hour24:      false,
			color:       graphic.Color{255, 255, 255},
			expected:    []byte{8, 0, 6, 1, 128, 255, 255, 255}, // 128 = date bit
		},
		{
			name:        "with 24 hour format",
			style:       ClockDefault,
			visibleDate: false,
			hour24:      true,
			color:       graphic.Color{255, 255, 255},
			expected:    []byte{8, 0, 6, 1, 64, 255, 255, 255}, // 64 = 24h bit
		},
		{
			name:        "with custom color",
			style:       ClockChristmas,
			visibleDate: true,
			hour24:      true,
			color:       graphic.Color{255, 0, 0},
			expected:    []byte{8, 0, 6, 1, 193, 255, 0, 0}, // 193 = 1 + 128 + 64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeviceConnectionMock{}
			err := SetClockMode(mock, tt.style, tt.visibleDate, tt.hour24, tt.color)
			require.NoError(t, err)
			require.Len(t, mock.WrittenPackets, 1)
			assert.Equal(t, tt.expected, mock.WrittenPackets[0])
		})
	}
}

func TestSetTime(t *testing.T) {
	tests := []struct {
		name                                            string
		year, month, day, weekDay, hour, minute, second int
		expected                                        []byte
	}{
		{
			name:     "sets time correctly",
			year:     24, month: 12, day: 25, weekDay: 3, hour: 14, minute: 30, second: 45,
			expected: []byte{11, 0, 1, 128, 24, 12, 25, 3, 14, 30, 45},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeviceConnectionMock{}
			err := SetTime(mock, tt.year, tt.month, tt.day, tt.weekDay, tt.hour, tt.minute, tt.second)
			require.NoError(t, err)
			require.Len(t, mock.WrittenPackets, 1)
			assert.Equal(t, tt.expected, mock.WrittenPackets[0])
		})
	}
}

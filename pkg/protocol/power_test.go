package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetPowerState(t *testing.T) {
	tests := []struct {
		name     string
		on       bool
		expected []byte
	}{
		{
			name:     "power on",
			on:       true,
			expected: []byte{5, 0, 7, 1, 1},
		},
		{
			name:     "power off",
			on:       false,
			expected: []byte{5, 0, 7, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeviceConnectionMock{}
			err := SetPowerState(mock, tt.on)
			require.NoError(t, err)
			require.Len(t, mock.WrittenPackets, 1)
			assert.Equal(t, tt.expected, mock.WrittenPackets[0])
		})
	}
}

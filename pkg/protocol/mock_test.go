package protocol

import "fmt"

// DeviceConnectionMock is a mock implementation of DeviceConnection for testing.
type DeviceConnectionMock struct {
	// Captured data
	WrittenPackets [][]byte
	DrainCalled    bool

	// Mock responses for ReadResponse()
	Responses     [][]byte
	responseIndex int

	// Error injection
	WritePacketErr error
	ReadErr        error
}

func (m *DeviceConnectionMock) WritePacket(packet []byte) error {
	if m.WritePacketErr != nil {
		return m.WritePacketErr
	}
	// Copy packet to avoid issues with slice reuse
	copied := make([]byte, len(packet))
	copy(copied, packet)
	m.WrittenPackets = append(m.WrittenPackets, copied)
	return nil
}

func (m *DeviceConnectionMock) ReadResponse() ([]byte, error) {
	if m.ReadErr != nil {
		return nil, m.ReadErr
	}
	if m.responseIndex >= len(m.Responses) {
		return nil, fmt.Errorf("no more mock responses available")
	}
	response := m.Responses[m.responseIndex]
	m.responseIndex++
	return response, nil
}

func (m *DeviceConnectionMock) DrainResponses() {
	m.DrainCalled = true
}

// AddResponse queues a response for ReadResponse to return.
func (m *DeviceConnectionMock) AddResponse(response []byte) {
	m.Responses = append(m.Responses, response)
}

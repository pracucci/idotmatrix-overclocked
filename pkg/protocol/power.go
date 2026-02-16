package protocol

// SetPowerState turns the display on or off.
func SetPowerState(d DeviceConnection, on bool) error {
	var state uint8
	if on {
		state = 1
	}
	return WriteData(d, []byte{5, 0, 7, 1, state})
}

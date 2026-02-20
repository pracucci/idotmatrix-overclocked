package protocol

// SetPowerState turns the display on or off.
func SetPowerState(d DeviceConnection, on bool) error {
	var state uint8
	if on {
		state = 1
	}
	return WriteData(d, []byte{5, 0, 7, 1, state})
}

// SetBrightness sets the display brightness (5-100 percent).
func SetBrightness(d DeviceConnection, percent uint8) error {
	if percent < 5 {
		percent = 5
	}
	if percent > 100 {
		percent = 100
	}
	return WriteData(d, []byte{5, 0, 4, 128, percent})
}

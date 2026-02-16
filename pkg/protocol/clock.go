package protocol

import "github.com/pracucci/idotmatrix-overclocked/pkg/graphic"

// Clock style constants
const (
	ClockDefault           = iota
	ClockChristmas         = iota
	ClockRacing            = iota
	ClockInverted          = iota
	ClockAnimatedHourGlass = iota
)

// SetClockMode sets the clock display mode on the device.
func SetClockMode(d DeviceConnection, style int, visibleDate bool, hour24 bool, color graphic.Color) error {
	var sb uint8 = uint8(style)
	if visibleDate {
		sb |= 128
	}
	if hour24 {
		sb |= 64
	}
	return WriteData(d, []byte{8, 0, 6, 1, sb, color[0], color[1], color[2]})
}

// SetTime sets the current time on the device.
func SetTime(d DeviceConnection, year int, month int, day int, weekDay int, hour int, minute int, second int) error {
	return WriteData(d, []byte{11, 0, 1, 128, byte(year), byte(month), byte(day), byte(weekDay), byte(hour), byte(minute), byte(second)})
}

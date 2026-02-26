package protocol

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"tinygo.org/x/bluetooth"
)

const serviceId = uint16(0x00fa)

var serviceUUID = bluetooth.New16BitUUID(serviceId)

const writeCharacteristicId = uint16(0xfa02)

var writeCharacteristicUUID = bluetooth.New16BitUUID(writeCharacteristicId)

const readCharacteristicId = uint16(0xfa03)

var readCharacteristicUUID = bluetooth.New16BitUUID(readCharacteristicId)

var btAdapter = bluetooth.DefaultAdapter

// DeviceNamePrefix is the prefix for iDotMatrix device names.
const DeviceNamePrefix = "IDM-"

// Device represents a connection to an iDotMatrix display.
type Device struct {
	logger              log.Logger
	scanResult          bluetooth.ScanResult
	btDevice            *bluetooth.Device
	writeCharacteristic bluetooth.DeviceCharacteristic
	readCharacteristic  bluetooth.DeviceCharacteristic
	responseChan        chan []byte
}

// NewDevice creates a new Device instance.
// Use Connect() to establish a connection to the device.
func NewDevice(logger log.Logger) *Device {
	return &Device{logger: logger}
}

// Connect establishes a BLE connection to an iDotMatrix device.
// If targetAddr is empty, it auto-discovers the first device with name prefix "IDM-".
// If targetAddr is specified, it connects to that specific MAC address.
func (d *Device) Connect(targetAddr string) error {
	if err := btAdapter.Enable(); err != nil {
		// Ignore "already enabled" errors during reconnection
		if !strings.Contains(err.Error(), "already") {
			return err
		}
	}

	// Scan for device
	if targetAddr == "" {
		level.Info(d.logger).Log("msg", "Scanning for iDotMatrix devices")
	}

	err := btAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		name := result.LocalName()
		level.Debug(d.logger).Log("msg", "Found device", "address", result.Address.String(), "rssi", result.RSSI, "name", name)

		if targetAddr != "" {
			// Connect to specific address
			if strings.EqualFold(result.Address.String(), targetAddr) {
				d.scanResult = result
				adapter.StopScan()
			}
		} else {
			// Auto-discover by name prefix
			if strings.HasPrefix(name, DeviceNamePrefix) {
				level.Info(d.logger).Log("msg", "Selected device", "name", name, "address", result.Address.String())
				d.scanResult = result
				adapter.StopScan()
			}
		}
	})
	if err != nil {
		return err
	}

	if d.scanResult.Address.String() == "" {
		if targetAddr != "" {
			return fmt.Errorf("device with address %q not found", targetAddr)
		}
		return fmt.Errorf("no iDotMatrix device found (looking for devices with name starting with %q)", DeviceNamePrefix)
	}

	btd, err := btAdapter.Connect(d.scanResult.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return err
	}

	srvcs, err := btd.DiscoverServices([]bluetooth.UUID{serviceUUID})
	if err != nil {
		return fmt.Errorf("service discover failed")
	}
	if len(srvcs) == 0 {
		return fmt.Errorf("device doesn't support %s service", serviceUUID.String())
	}

	service := srvcs[0]

	if !service.Is16Bit() || service.UUID().Get16Bit() != serviceId {
		return fmt.Errorf("invalid service id")
	}

	chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{writeCharacteristicUUID, readCharacteristicUUID})
	if err != nil {
		return err
	}
	if len(chars) != 2 {
		return fmt.Errorf("unexpected number of characteristics. expected 2, got %d", len(chars))
	}

	for _, ch := range chars {
		if !ch.Is16Bit() {
			return fmt.Errorf("invalid char type")
		}
		switch ch.Get16Bit() {
		case writeCharacteristicId:
			d.writeCharacteristic = ch
		case readCharacteristicId:
			d.readCharacteristic = ch
		default:
			return fmt.Errorf("invalid characteristic %s", ch.UUID().String())
		}
	}

	d.btDevice = btd

	// Set up response channel and enable notifications
	d.responseChan = make(chan []byte, 1)
	err = d.readCharacteristic.EnableNotifications(func(buf []byte) {
		// Copy the buffer since it may be reused
		data := make([]byte, len(buf))
		copy(data, buf)
		// Non-blocking send to channel (drop if channel is full)
		select {
		case d.responseChan <- data:
		default:
		}
	})
	if err != nil {
		fmt.Printf("Warning: could not enable notifications: %v\n", err)
	}

	return nil
}

// Disconnect closes the BLE connection to the device.
func (d *Device) Disconnect() error {
	if d.btDevice == nil {
		return nil
	}
	return d.btDevice.Disconnect()
}

// ReadResponse waits for a response from the device via BLE notifications.
// This is used after sending data chunks to wait for the device to process them.
// Returns the response bytes and any error. Times out after 2 seconds.
func (d *Device) ReadResponse() ([]byte, error) {
	if d.responseChan == nil {
		return nil, fmt.Errorf("notifications not enabled")
	}
	select {
	case data := <-d.responseChan:
		return data, nil
	case <-time.After(2 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// WritePacket writes a single BLE packet without internal chunking.
// Used for precise control over packet boundaries.
func (d *Device) WritePacket(packet []byte) error {
	_, err := d.writeCharacteristic.WriteWithoutResponse(packet)
	return err
}

// DrainResponses clears any stale notifications from the response channel.
// Call this before starting a new multi-chunk upload to ensure clean state.
func (d *Device) DrainResponses() {
	if d.responseChan == nil {
		return
	}
	for {
		select {
		case <-d.responseChan:
			// Discard stale data
		default:
			return
		}
	}
}

// DeviceConnection is the interface for communicating with an iDotMatrix device.
// This interface is implemented by Device and allows protocol functions
// to be decoupled from the concrete device implementation.
type DeviceConnection interface {
	// WritePacket writes a single BLE packet without internal chunking.
	WritePacket(packet []byte) error

	// ReadResponse waits for a response from the device via BLE notifications.
	ReadResponse() ([]byte, error)

	// DrainResponses clears any stale notifications from the response channel.
	DrainResponses()
}

// WriteData writes data to the device in up to MTU sized chunks.
// The iDotMatrix device has a 514-byte MTU limit.
func WriteData(d DeviceConnection, data []byte) error {
	const maxMTU = 514

	cursor := 0
	remaining := len(data)
	for remaining > 0 {
		chunkLen := min(maxMTU, remaining)
		if err := d.WritePacket(data[cursor : cursor+chunkLen]); err != nil {
			return err
		}
		cursor += chunkLen
		remaining -= chunkLen
	}

	return nil
}

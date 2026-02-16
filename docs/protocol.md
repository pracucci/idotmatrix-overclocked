# iDot Device Protocol Specification

This document describes the BLE communication protocol for iDot pixel display devices.

## BLE Communication

### Service and Characteristics

| Component | UUID (16-bit) | Full UUID |
|-----------|---------------|-----------|
| Service | `0x00fa` | `000000fa-0000-1000-8000-00805f9b34fb` |
| Write Characteristic | `0xfa02` | `0000fa02-0000-1000-8000-00805f9b34fb` |
| Read Characteristic | `0xfa03` | `0000fa03-0000-1000-8000-00805f9b34fb` |

- **Write**: Uses WriteWithoutResponse mode
- **Read**: Uses notifications (must be enabled after connection)
- **Write MTU**: 514 bytes per BLE packet

### Connection Sequence

1. Scan for devices advertising service UUID `0x00fa`
2. Connect to device
3. Wait 100ms for BLE stack to stabilize
4. Enable notifications on read characteristic (`0xfa03`)
5. Ready for commands

---

## Static Image Transmission

Static images are sent as raw RGB data with a 9-byte header per chunk.

### Header Format (9 bytes, little-endian)

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0 | 2 | Packet Length | Chunk data length + 9 |
| 2 | 1 | Command | `0x00` (image data) |
| 3 | 1 | Subcommand | `0x00` |
| 4 | 1 | Packet Type | `0x00` = first, `0x02` = continuation |
| 5 | 4 | Total Size | Total image size in bytes |

### Image Format

- **Dimensions**: 64x64 or 32x32 pixels
- **Color depth**: 24-bit RGB (3 bytes per pixel)
- **Byte order**: R, G, B per pixel, row-major (top-to-bottom, left-to-right)
- **Total size**: 12,288 bytes (64x64) or 3,072 bytes (32x32)

### Transmission Flow

1. Send `SetDrawMode(1)` command
2. Split image into 4096-byte chunks
3. For each chunk:
   - Build 9-byte header
   - Append chunk data
   - Write to device (BLE stack handles MTU splitting)
4. Wait 500ms before disconnect

### Pixel Offset Calculation

```
offset = (y * width + x) * 3
pixel[offset+0] = R
pixel[offset+1] = G
pixel[offset+2] = B
```

---

## GIF Transmission

Animated GIFs use a 16-byte header with CRC32-based caching.

### Header Format (16 bytes, little-endian)

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0 | 2 | Packet Length | Chunk data length + 16 |
| 2 | 1 | Command | `0x01` (GIF data) |
| 3 | 1 | Subcommand | `0x00` |
| 4 | 1 | Packet Type | `0x00` = first, `0x02` = continuation |
| 5 | 4 | Total Size | Total GIF size in bytes |
| 9 | 4 | CRC32 | IEEE CRC32 of entire GIF |
| 13 | 2 | Time Signature | `0x0000` (reserved) |
| 15 | 1 | GIF Type | `0x0C` (12 = no time signature) |

### Transmission Flow

1. Calculate CRC32 (IEEE) of entire GIF data
2. Split GIF into 4096-byte chunks
3. For each chunk:
   - Build 16-byte header
   - Append chunk data
   - Split into 509-byte BLE packets
   - Send each packet with 10ms delay between packets
   - Read response after all packets for chunk are sent
4. Process response

### Response Format

Expected: `[0x05, 0x00, 0x01, 0x00, status]`

| Status | Meaning | Action |
|--------|---------|--------|
| `0x01` | Chunk OK | Continue to next chunk |
| `0x03` | Complete/Cached | Stop sending (success) |

### GIF Caching

The device caches previously uploaded GIFs indexed by CRC32:

- When first chunk is sent, device checks CRC32 against cache
- If found (cache hit): Returns `0x03` immediately
- Client should stop sending and return success
- Subsequent plays of the same GIF skip the upload entirely

This optimization significantly reduces bandwidth for repeated animations.

### GIF Requirements

- Dimensions: 64x64 pixels
- Format: Standard GIF with proper disposal methods
- Disposal: Set to DisposalBackground (0x02) for all frames
- Loop count: 0 (infinite loop)

---

## Device Commands

### SetDrawMode

Switches device to image reception mode. Required before sending static images.

```
Packet: [0x05, 0x00, 0x04, 0x01, mode]
```

| Mode | Description |
|------|-------------|
| 0 | Default/display mode |
| 1 | Image reception mode |

### SetPixel (Graffiti Mode)

Draws a single pixel at specified coordinates.

```
Packet: [0x0A, 0x00, 0x05, 0x01, 0x00, R, G, B, X, Y]
```

| Byte | Field | Range |
|------|-------|-------|
| 5 | Red | 0-255 |
| 6 | Green | 0-255 |
| 7 | Blue | 0-255 |
| 8 | X | 0-63 |
| 9 | Y | 0-63 |

Coordinates: (0,0) = top-left, (63,63) = bottom-right

### SetPowerState

Turns the display on or off.

```
Packet: [0x05, 0x00, 0x07, 0x01, state]
```

| State | Description |
|-------|-------------|
| 0x00 | Power off |
| 0x01 | Power on |

### SetClockMode

Displays a clock with configurable style and color.

```
Packet: [0x08, 0x00, 0x06, 0x01, style, R, G, B]
```

**Style byte encoding:**
- Bits 0-5: Clock style (0-4)
- Bit 6 (0x40): 24-hour format if set
- Bit 7 (0x80): Show date if set

| Style | Description |
|-------|-------------|
| 0 | Default |
| 1 | Christmas |
| 2 | Racing |
| 3 | Inverted |
| 4 | Animated hourglass |

### SetTime

Sets the device's internal clock.

```
Packet: [0x0B, 0x00, 0x01, 0x80, year, month, day, weekday, hour, minute, second]
```

| Byte | Field | Range |
|------|-------|-------|
| 4 | Year | 0-99 (2-digit) |
| 5 | Month | 1-12 |
| 6 | Day | 1-31 |
| 7 | Weekday | 0-6 (0=Monday) |
| 8 | Hour | 0-23 |
| 9 | Minute | 0-59 |
| 10 | Second | 0-59 |

---

## Timing Requirements

| Parameter | Value | Purpose |
|-----------|-------|---------|
| Post-connection delay | 100ms | BLE stack stabilization |
| GIF inter-packet delay | 10ms | Device processing time |
| Response timeout | 2s | Maximum wait for device response |
| Post-transmission delay | 500ms | Ensure final writes complete |
| Pixel batch delay | 5ms | Between individual SetPixel calls |

---

## Constants Summary

```go
// BLE
ServiceUUID              = 0x00fa
WriteCharacteristic      = 0xfa02
ReadCharacteristic       = 0xfa03
WriteMTU                 = 514

// Static Image
ImageHeaderSize          = 9
ImageChunkSize           = 4096
ImageCommand             = 0x00
PacketTypeFirst          = 0x00
PacketTypeContinuation   = 0x02

// GIF
GIFHeaderSize            = 16
GIFChunkSize             = 4096
GIFBLEPacketSize         = 509
GIFCommand               = 0x01
GIFTypeNoTimestamp       = 0x0C

// Response
ResponseContinue         = 0x01
ResponseComplete         = 0x03

// Display
DisplayWidth             = 64
DisplayHeight            = 64
BufferSize               = 12288  // 64 * 64 * 3
```

# Code Structure Documentation

## AI Maintenance Instructions

When making changes to this codebase, AI agents should update this document to reflect:

1. **New files or directories** - Add to the directory structure and describe their purpose
2. **New types or structs** - Add to the Core Types section with field descriptions
3. **New constants** - Add to the Constants section
4. **Changed data flows** - Update the Data Flow section if protocols change
5. **New CLI commands** - Add to the cmd/ package description and Key Files table

Keep descriptions concise and focused on helping developers navigate quickly.

---

## Directory Structure

```
idm-cli/
├── Makefile                   # Build targets
├── go.mod                     # Module definition (Go 1.24.0)
├── go.sum                     # Dependency lock file
├── cmd/
│   └── cli/                   # CLI commands (Cobra)
│       ├── main.go            # CLI entry point and root command
│       ├── discover.go        # Bluetooth device scanner
│       ├── fire.go            # DOOM-style fire animation
│       ├── clock.go           # Digital clock display
│       ├── server.go          # HTTP server command
│       ├── showgif.go         # GIF file display
│       ├── showimage.go       # Static image display
│       ├── text.go            # Text rendering with animations
│       └── snake.go           # Snake game
├── idot/                      # BLE device abstraction
│   ├── doc.go                 # Package documentation
│   └── device.go              # BLE connection & communication
├── pkg/graphic/               # Graphics utilities (colors, images, buffers)
│   ├── color.go               # Color type, palette, shadows
│   ├── image.go               # Image container types, display constants
│   └── image_test.go          # Tests for image and color functions
├── pkg/protocol/              # iDotMatrix communication protocol
│   ├── device.go              # DeviceConnection interface
│   ├── clock.go               # Clock display modes
│   ├── gif.go                 # Animated GIF protocol
│   ├── graffiti.go            # Individual pixel setting
│   └── image.go               # Static image protocol
├── pkg/text/                  # Text rendering package
│   ├── text.go                # Text layout, wrapping, multi-line centering
│   ├── animation.go           # Text animation generation
│   ├── draw.go                # Low-level pixel drawing
│   └── font.go                # 5x7 bitmap font
├── pkg/games/snake/           # Snake game implementation
│   ├── game.go                # Game logic
│   ├── interstitial.go        # Level transition animations
│   ├── level.go               # Level definitions
│   ├── map.go                 # Game map
│   └── render.go              # Game rendering
├── pkg/server/                # HTTP server for web control
│   ├── server.go              # Server struct, device management, graceful shutdown
│   ├── api.go                 # API endpoint handlers (/api/*)
│   ├── console.go             # Web console static file serving
│   └── assets/                # Embedded web console assets
│       ├── index.html         # Main HTML page
│       ├── style.css          # Dark theme styles
│       └── app.js             # UI logic, API calls, preview canvas
├── testdata/                  # Test assets
│   ├── demo.gif
│   ├── test_64x64.gif
│   └── test_64x64.png
└── docs/                      # Documentation
    └── structure.md           # This file
```

---

## Package Organization

### `idot/` - BLE Device Abstraction

Abstracts Bluetooth Low Energy communication with iDot matrix displays.

| File | Purpose |
|------|---------|
| `device.go` | BLE device discovery, connection, characteristic negotiation, packet writing |

### `pkg/graphic/` - Graphics Utilities

Core graphics types including colors, images, and display buffer utilities.

| File | Purpose |
|------|---------|
| `color.go` | `Color` type, color palette, shadow colors, `ShadowFor()` function |
| `image.go` | `Image` struct, display constants, buffer creation, pixel setting |

### `pkg/protocol/` - Communication Protocol

Protocol packet construction and encoding for iDot matrix displays.

| File | Purpose |
|------|---------|
| `device.go` | `DeviceConnection` interface for device abstraction |
| `clock.go` | `SetClockMode()`, `SetTime()`, clock style constants |
| `image.go` | `SetDrawMode()`, `SendImage()` for RGB data (4096-byte chunks, 9-byte headers) |
| `gif.go` | `SendGIF()` for animated GIFs (4096-byte chunks, 16-byte headers, CRC32) |
| `graffiti.go` | `SetPixel()` for individual pixel updates |

### `pkg/text/` - Text Rendering

Generates static and animated text images for the 64x64 display.

| File | Purpose |
|------|---------|
| `text.go` | Text layout, wrapping, multi-line centering |
| `animation.go` | GIF-based animations (blink, appear, disappear) |
| `draw.go` | Low-level pixel and character rendering |
| `font.go` | 5x7 bitmap font data and text width calculations |

### `cmd/` - CLI Commands

Cobra-based CLI providing end-user functionality.

| Command | Purpose |
|---------|---------|
| `discover` | Discover nearby Bluetooth devices |
| `text` | Display text with optional animations |
| `showimage` | Display static PNG/JPEG/GIF images |
| `showgif` | Display animated GIFs with frame optimization |
| `clock` | Configure and display digital clock |
| `fire` | Generate DOOM-style fire animation |
| `snake` | Interactive snake game |
| `server` | HTTP server with API and web console |

### `pkg/server/` - HTTP Server

Web-based control interface for the iDotMatrix display.

| File | Purpose |
|------|---------|
| `server.go` | Server struct, device connection management, HTTP setup, graceful shutdown |
| `api.go` | API endpoint handlers for all display commands |
| `console.go` | Serves embedded static files for web console |
| `assets/*` | HTML, CSS, JS for web console UI |

---

## Key Files Quick Reference

| Feature | Primary File | Supporting Files |
|---------|--------------|------------------|
| BLE Connection | `idot/device.go` | - |
| Send Static Image | `pkg/protocol/image.go` | `idot/device.go` |
| Send Animated GIF | `pkg/protocol/gif.go` | `idot/device.go` |
| Set Individual Pixel | `pkg/protocol/graffiti.go` | `idot/device.go` |
| Clock Display | `pkg/protocol/clock.go` | `idot/device.go` |
| Color Palette | `pkg/graphic/color.go` | - |
| Image Buffers | `pkg/graphic/image.go` | - |
| Text Layout | `pkg/text/text.go` | `pkg/text/font.go` |
| Text Animations | `pkg/text/animation.go` | `pkg/text/draw.go` |
| Character Drawing | `pkg/text/draw.go` | `pkg/text/font.go` |

---

## Core Types

### `idot.Device`

BLE device connection and communication.

```go
type Device struct {
    scanResult          bluetooth.ScanResult
    btDevice            *bluetooth.Device
    writeCharacteristic bluetooth.DeviceCharacteristic
    writeMTU            int
    readCharacteristic  bluetooth.DeviceCharacteristic
    readMTU             int
    responseChan        chan []byte
}
```

**Key Methods:**
- `NewDevice(targetAddr string)` - Discover device by MAC address
- `DiscoverDevice()` - Auto-discover first iDotMatrix device
- `Connect()` - Establish BLE connection
- `Disconnect()` - Close connection
- `Write(packet []byte)` - Write packet with MTU chunking
- `WriteBLEPacket(packet []byte)` - Write single BLE packet
- `ReadResponse()` - Wait for device response
- `DrainResponses()` - Clear stale notifications

### `protocol.DeviceConnection`

Interface for writing protocol packets to a device.

```go
type DeviceConnection interface {
    Write(packet []byte) error
    WriteBLEPacket(packet []byte) error
    ReadResponse() ([]byte, error)
    DrainResponses()
}
```

### `graphic.Color`

RGB color as byte array.

```go
type Color [3]uint8

// Predefined colors available:
// White, Black, Red, Green, Blue, Yellow, Cyan, Magenta,
// Orange, Gray, Purple, Pink, and Dark variants for shadows
```

### `graphic.Image`

Container for static or animated images.

```go
type Image struct {
    Type       ImageType  // Static or Animated
    StaticData []byte     // Raw RGB data (64*64*3 bytes)
    GIFData    *gif.GIF   // Animated GIF
}
```

### `text.TextOptions`

Configuration for text rendering.

```go
type TextOptions struct {
    TextColor   graphic.Color  // Main text color
    ShadowColor graphic.Color  // Shadow color
    Background  graphic.Color  // Background fill
    ShadowX     int            // Shadow X offset (default: 1)
    ShadowY     int            // Shadow Y offset (default: 1)
}
```

### `text.AnimationOptions`

Extended options for animated text.

```go
type AnimationOptions struct {
    TextOptions
    FrameDelay    int  // Delay per frame (10ms units)
    BlinkOffDelay int  // Off-frame delay for blink
    LetterDelay   int  // Delay between appearing letters
    HoldDelay     int  // Final frame hold delay
}
```

---

## Constants

### Display Dimensions (`pkg/graphic/image.go`)

```go
const (
    DisplayWidth  = 64
    DisplayHeight = 64
    BufferSize    = DisplayWidth * DisplayHeight * 3  // 12,288 bytes
)
```

### Font Metrics (`pkg/text/font.go`)

```go
const (
    FontWidth   = 5   // Character width in pixels
    FontHeight  = 7   // Character height in pixels
    FontSpacing = 6   // Horizontal spacing (includes 1px gap)
    LineSpacing = 4   // Pixels between lines
)
```

### BLE UUIDs (`idot/device.go`)

```go
const iDotServiceId              = uint16(0x00fa)
const iDotWriteCharacteristicId  = uint16(0xfa02)
const iDotReadCharacteristicId   = uint16(0xfa03)
```

### Protocol Constants

```go
// Image protocol (pkg/protocol/image.go)
const headerSize = 9  // Bytes per image packet header

// GIF protocol (pkg/protocol/gif.go)
const gifHeaderSize    = 16
const gifChunkSize     = 4096
const gifBLEPacketSize = 509

// Clock styles (pkg/protocol/clock.go)
const ClockDefault           = 0
const ClockChristmas         = 1
const ClockRacing            = 2
const ClockInverted          = 3
const ClockAnimatedHourGlass = 4

// Limits (cmd/cli/showgif.go)
const maxFrames      = 64
const maxDurationMs  = 2000
const minFrameTimeMs = 16  // ~60 FPS limit
```

---

## Data Flow

### Static Text → Display

```
User Input (text + colors)
    ↓
text.GenerateStaticText()
    ├── WrapText() - Split to lines
    ├── DrawMultiLineCentered() - Layout
    └── DrawTextShadowed() - Render pixels
    ↓
graphic.Image (StaticData = 12,288 bytes RGB)
    ↓
protocol.SetDrawMode(device, 1)
    ↓
protocol.SendImage(device, data)
    ├── Split into 4096-byte chunks
    ├── Add 9-byte header per chunk
    └── Send as 514-byte BLE packets
    ↓
iDot Display
```

### Animated Text → Display

```
User Input (text + animation type)
    ↓
text.GenerateBlinkingText() / GenerateAppearingText() / etc.
    ├── Create multiple frame buffers
    └── Assemble into gif.GIF
    ↓
graphic.Image (GIFData = *gif.GIF)
    ↓
protocol.SendGIF(device, gifBytes)
    ├── Calculate CRC32
    ├── Split into 4096-byte chunks
    ├── Add 16-byte header (with CRC32)
    ├── Send as 509-byte BLE packets
    └── Wait for acknowledgment per chunk
    ↓
iDot Display (animated playback)
```

### Image File → Display

```
Image File (PNG/JPEG/GIF)
    ↓
image.Decode() → Go image.Image
    ↓
Validate 64x64 dimensions
    ↓
Convert to RGB (12,288 bytes)
    ↓
[Same as Static Text from SendImage]
```

---

## Common Patterns

### Device Connection

```go
device, err := idot.NewDevice(targetAddr)
if err != nil {
    return err
}

if err := device.Connect(); err != nil {
    return err
}
defer device.Disconnect()

// ... send content ...

time.Sleep(500 * time.Millisecond)  // Allow final writes
```

### Static Content

```go
protocol.SetDrawMode(device, 1)
protocol.SendImage(device, imageData)
```

### Animated Content

```go
gifBytes, _ := img.GIFBytes()
protocol.SendGIF(device, gifBytes)
```

### Text Generation

```go
opts := text.DefaultTextOptions()
opts.TextColor = graphic.Red
opts.ShadowColor = graphic.ShadowFor(graphic.Red)

lines := text.WrapText(message)
if text.TextBlockHeight(lines) > graphic.DisplayHeight {
    return fmt.Errorf("text too tall")
}

img := text.GenerateStaticText(message, opts)
rawBytes, _ := img.RawBytes()
protocol.SetDrawMode(device, 1)
protocol.SendImage(device, rawBytes)
```

### Animation Options Defaults

```go
text.DefaultAnimationOptions() returns:
    FrameDelay:    50   // 500ms
    BlinkOffDelay: 30   // 300ms
    LetterDelay:   20   // 200ms
    HoldDelay:     100  // 1 second
```

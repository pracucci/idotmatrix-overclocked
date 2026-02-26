# AGENT.md

Instructions for AI coding agents working on this codebase.

## Project Overview

**idm-cli** is a Go CLI application for controlling iDot 64x64 pixel matrix displays via Bluetooth Low Energy (BLE). It supports displaying static images, animated GIFs, text with animations, a digital clock, and interactive games.

## Quick Start

```bash
# Build the project
make build

# Run a command (requires device MAC address)
./idm-cli text --target XX:XX:XX:XX:XX:XX --text "Hello"
```

## Code Structure

See `docs/structure.md` for detailed documentation including:
- Directory structure and package organization
- Core types and their purposes
- Data flow diagrams
- Common patterns

**Keep `docs/structure.md` updated when making structural changes.**

## Key Packages

| Package | Purpose |
|---------|---------|
| `idot/` | BLE device abstraction (connection, packet writing) |
| `pkg/protocol/` | iDotMatrix communication protocol (packet construction) |
| `pkg/text/` | Text rendering, fonts, animations |
| `pkg/server/` | HTTP server with API and web console |
| `cmd/` | CLI commands (Cobra framework) |

## Development Guidelines

### Logger Conventions

When passing a logger to a function, always make it the last argument.

### Testing Requirements

All changes must pass tests before completion:
```bash
go test ./...
```

When modifying `pkg/img` or `pkg/text`, ensure existing tests pass and add tests for new functionality.

### Adding a New CLI Command

1. Create a new file in `cmd/cli/` with the command name
2. Implement the Cobra command with a unique exported variable name (e.g., `FooCmd`)
3. Register the command in `cmd/cli/main.go` using `init()`
4. Follow existing command patterns (see `cmd/cli/text.go`)

### Adding Device Functionality

1. Add protocol implementation in `pkg/protocol/` package
2. Document packet structure in code comments
3. Follow existing patterns for chunking and BLE writes
4. Use the `DeviceConnection` interface for device abstraction

### Adding API Endpoints

When adding a new display command that should be accessible via the HTTP server:

1. Add the endpoint handler in `pkg/server/api.go`:
   - Register the route in `registerAPIRoutes()`
   - Create a handler method on `*Server` (e.g., `handleNewCommand`)
   - Use `s.withDevice()` to serialize device access
   - Return JSON via `writeJSON(w, apiResponse{...})`

2. Add the UI in `pkg/server/assets/`:
   - Add form fields/card in `index.html`
   - Add action handler in `app.js` `handleAction()` switch
   - Style as needed in `style.css`

3. API endpoint conventions:
   - Base path: `/api/`
   - Methods: POST for actions, GET for status
   - Query parameters for simple inputs
   - Multipart form for file uploads
   - Response format: `{"success": true}` or `{"success": false, "error": "message"}`

Example endpoint handler:
```go
func (s *Server) handleNewCommand(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeJSON(w, apiResponse{Success: false, Error: "method not allowed"})
        return
    }

    param := r.URL.Query().Get("param")
    if param == "" {
        writeJSON(w, apiResponse{Success: false, Error: "param required"})
        return
    }

    err := s.withDevice(func(d *protocol.Device) error {
        // ... send to device ...
        time.Sleep(500 * time.Millisecond)
        return nil
    })

    if err != nil {
        writeJSON(w, apiResponse{Success: false, Error: err.Error()})
        return
    }

    writeJSON(w, apiResponse{Success: true})
}
```

### Text/Image Generation

1. Use `pkg/text/` for generating display content
2. Use `text.DefaultTextOptions()` or `text.DefaultAnimationOptions()` as starting points
3. Respect display constraints: 64x64 pixels, 5x7 font

## Constraints

- **Display size**: 64x64 pixels
- **Font**: 5x7 bitmap characters, 6px horizontal spacing
- **Image buffer**: 12,288 bytes (64 * 64 * 3 RGB)
- **BLE MTU**: 514 bytes for images, 509 bytes for GIFs
- **Max GIF frames**: 64
- **Max GIF duration**: 2000ms

## Testing

```bash
go test ./...
```

## Dependencies

- `tinygo.org/x/bluetooth` - BLE communication
- `github.com/spf13/cobra` - CLI framework
- Standard library `image`, `image/gif`, `image/png`

## Common Tasks

### Display static text
```go
opts := text.DefaultTextOptions()
img := text.GenerateStaticText("Hello", opts)
rawBytes, _ := img.RawBytes()
protocol.SetDrawMode(device, 1)
protocol.SendImage(device, rawBytes)
```

### Display animated text
```go
opts := text.DefaultAnimationOptions()
img := text.GenerateBlinkingText("Hello", opts)
gifBytes, _ := img.GIFBytes()
protocol.SendGIF(device, gifBytes)
```

### Connect to device
```go
device, _ := idot.NewDevice(macAddress)
device.Connect()
defer device.Disconnect()
// ... use device ...
time.Sleep(500 * time.Millisecond)
```

## Files to Update

When making changes, update these files as needed:

| Change Type | Files to Update |
|-------------|-----------------|
| New CLI command | `cmd/cli/<name>.go`, `cmd/cli/main.go`, `docs/structure.md` |
| New device protocol | `pkg/protocol/<protocol>.go`, `docs/structure.md` |
| New text feature | `pkg/text/*.go`, `docs/structure.md` |
| New constants | Relevant package file, `docs/structure.md` |
| New types | Relevant package file, `docs/structure.md` |
| CLI command added/removed/modified | `README.md` (CLI Commands section) |
| New API endpoint | `pkg/server/api.go`, `pkg/server/assets/*`, `docs/structure.md` |

## Preview GIF Regeneration

Preview GIFs for the README are generated by `tools/preview-generator`.

### Running

```bash
go run ./tools/preview-generator
# or
make generate-previews
```

### When to Regenerate

- Snake rendering changes (`pkg/games/snake/render.go`)
- Text animation changes (`pkg/text/animation.go`)
- New emoji/grot assets added

**ALWAYS ASK USER** before regenerating snake preview after significant snake logic changes.

### What Snake Preview Shows

1. Cover screen with "SNAKE" title (2s)
2. "LEVEL 1" appearing letter-by-letter
3. Gameplay: snake moving right on terrain (2s)

### Key Files Affecting Snake Preview

- `pkg/games/snake/render.go`: `GenerateCoverImage()`, `GenerateBackgroundWithObstacles()`
- `pkg/games/snake/interstitial.go`: Level text animation style

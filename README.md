# iDotMatrix overclocked

## Requirements

- iDotMatrix 64x64 pixel display

## Building

```bash
make build
```

## Auto-Discovery

All commands support automatic device discovery. If `--target` is not specified, the tool will scan for nearby iDotMatrix devices (names starting with "IDM-") and connect to the first one found (sorted alphabetically).

## CLI Commands

### snake

Play Snake on the iDot display with progressive difficulty levels.

```bash
./idm-cli snake
```

Options:
- `--target`: Bluetooth MAC address of the display (auto-discovers if not specified)
- `--level`: Starting level (default: 1)

Controls: WASD or Arrow keys to move, Q to quit

### text

Display text on the iDot display with optional animations.

```bash
./idm-cli text --text "HELLO"
```

Options:
- `--target`: Bluetooth MAC address of the display (auto-discovers if not specified)
- `--text` (required): Text to display (uppercase A-Z, 0-9, punctuation)
- `--animation`: Animation type (none, blink, appear, appear-disappear)
- `--color`: Text color (white, red, green, blue, yellow, etc.)
- `--verbose`: Enable verbose debug logging

### fire

Generate and display a DOOM-style fire animation.

```bash
./idm-cli fire
```

Options:
- `--target`: Bluetooth MAC address of the display (auto-discovers if not specified)
- `--verbose`: Enable verbose debug logging

### clock

Show and configure the clock on the iDot display.

```bash
./idm-cli clock
```

Options:
- `--target`: Bluetooth MAC address of the display (auto-discovers if not specified)
- `--time`: Time value in RFC1123Z format (defaults to current time)
- `--style`: Clock style (0:Default, 1:Christmas, 2:Racing, 3:Inverted, 4:Hour Glass)
- `--show-date`: Show date as well as time (default: true)
- `--24hour`: Show time in 24-hour format (default: true)
- `--color`: Clock color (default: white)
- `--verbose`: Enable verbose debug logging

### discover

Discover nearby Bluetooth devices.

```bash
./idm-cli discover
```

Options:
- `--scan-time`: Max number of seconds to perform scan. 0 means infinite
- `--verbose`: Verbose output during scan

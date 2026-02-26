package snake

import (
	"os"

	"golang.org/x/term"
)

// InputSource provides key inputs to the game.
type InputSource interface {
	// ReadKey blocks until a key is pressed and returns it.
	ReadKey() rune
	// TryReadKey returns a key if available, or false if no key is waiting.
	TryReadKey() (rune, bool)
	// Close cleans up resources.
	Close()
}

// TerminalInput provides keyboard input from the terminal.
type TerminalInput struct {
	keys     chan rune
	oldState *term.State
	done     chan struct{}
}

// NewTerminalInput creates a new terminal input source.
func NewTerminalInput() (*TerminalInput, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	t := &TerminalInput{
		keys:     make(chan rune, 10),
		oldState: oldState,
		done:     make(chan struct{}),
	}

	go t.readLoop()
	return t, nil
}

func (t *TerminalInput) readLoop() {
	buf := make([]byte, 3)
	for {
		select {
		case <-t.done:
			return
		default:
		}

		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		if n > 0 {
			// Handle escape sequences for arrow keys
			if n == 3 && buf[0] == 27 && buf[1] == 91 {
				switch buf[2] {
				case 65: // up arrow
					t.keys <- 'w'
				case 66: // down arrow
					t.keys <- 's'
				case 67: // right arrow
					t.keys <- 'd'
				case 68: // left arrow
					t.keys <- 'a'
				}
			} else {
				t.keys <- rune(buf[0])
			}
		}
	}
}

// ReadKey blocks until a key is pressed.
func (t *TerminalInput) ReadKey() rune {
	return <-t.keys
}

// TryReadKey returns a key if available, or false if no key is waiting.
func (t *TerminalInput) TryReadKey() (rune, bool) {
	select {
	case key := <-t.keys:
		return key, true
	default:
		return 0, false
	}
}

// Close restores the terminal state and stops the input reader.
func (t *TerminalInput) Close() {
	close(t.done)
	if t.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), t.oldState)
	}
}

// ChannelInput provides key input via a channel (for HTTP-based input).
type ChannelInput struct {
	keys   chan rune
	closed bool
}

// NewChannelInput creates a new channel-based input source.
func NewChannelInput() *ChannelInput {
	return &ChannelInput{
		keys: make(chan rune, 10),
	}
}

// SendKey sends a key to the input channel.
func (c *ChannelInput) SendKey(key rune) {
	if c.closed {
		return
	}
	select {
	case c.keys <- key:
	default:
		// Channel full, drop the key
	}
}

// ReadKey blocks until a key is received.
func (c *ChannelInput) ReadKey() rune {
	key, ok := <-c.keys
	if !ok {
		return 'q' // Return quit if channel closed
	}
	return key
}

// TryReadKey returns a key if available, or false if no key is waiting.
func (c *ChannelInput) TryReadKey() (rune, bool) {
	select {
	case key := <-c.keys:
		return key, true
	default:
		return 0, false
	}
}

// Close closes the input channel.
func (c *ChannelInput) Close() {
	if !c.closed {
		c.closed = true
		close(c.keys)
	}
}

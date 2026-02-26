package server

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/pracucci/idotmatrix-overclocked/pkg/games/snake"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
)

// SnakeSession manages an active snake game session.
type SnakeSession struct {
	ID    string
	game  *snake.Game
	input *snake.ChannelInput
	done  chan struct{}
}

// SnakeManager manages snake game sessions (only one at a time).
type SnakeManager struct {
	mu      sync.Mutex
	session *SnakeSession
}

// NewSnakeManager creates a new snake manager.
func NewSnakeManager() *SnakeManager {
	return &SnakeManager{}
}

// generateSessionID generates a random session ID.
func generateSessionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// StartSession starts a new snake game session.
// Returns the session ID or an error if a session is already active.
func (m *SnakeManager) StartSession(device *protocol.Device) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if a session is already active
	if m.session != nil && m.session.game.IsRunning() {
		return m.session.ID, nil
	}

	// Clean up any old session
	if m.session != nil {
		m.session.input.Close()
		m.session = nil
	}

	// Create new session
	sessionID := generateSessionID()
	input := snake.NewChannelInput()
	game := snake.NewGame(device, 1)
	done := make(chan struct{})

	session := &SnakeSession{
		ID:    sessionID,
		game:  game,
		input: input,
		done:  done,
	}
	m.session = session

	// Start the game in a goroutine
	go func() {
		defer close(done)
		game.RunWithChannelInput(input)
	}()

	return sessionID, nil
}

// StopSession stops the active snake game session.
func (m *SnakeManager) StopSession() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.session == nil {
		return
	}

	// Stop the game
	m.session.game.Stop()
	m.session.input.Close()

	// Wait for the game to finish (with timeout via select not needed since Stop sets running=false)
	<-m.session.done

	m.session = nil
}

// SendInput sends a key input to the active game session.
// Returns the current game state.
func (m *SnakeManager) SendInput(key string) (snake.GameState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.session == nil || !m.session.game.IsRunning() {
		return snake.StateEnded, false
	}

	m.session.game.SendInput(key)
	return m.session.game.GetState(), true
}

// GetState returns the current game state.
func (m *SnakeManager) GetState() (snake.GameState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.session == nil {
		return snake.StateEnded, false
	}

	return m.session.game.GetState(), m.session.game.IsRunning()
}

// HasActiveSession returns whether there's an active game session.
func (m *SnakeManager) HasActiveSession() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.session != nil && m.session.game.IsRunning()
}

// GetSessionID returns the current session ID, or empty if no session.
func (m *SnakeManager) GetSessionID() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.session == nil {
		return ""
	}
	return m.session.ID
}

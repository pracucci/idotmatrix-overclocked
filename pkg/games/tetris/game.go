package tetris

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
)

// Game timing constants
const (
	RenderInterval = 100 * time.Millisecond // How often to render
	DropInterval   = 800 * time.Millisecond // Gravity speed (piece drops every interval)
)

// RandSource is an interface for random number generation (for testing)
type RandSource interface {
	Intn(n int) int
}

// defaultRand wraps math/rand for production use
type defaultRand struct{}

func (defaultRand) Intn(n int) int { return rand.Intn(n) }

// GameState contains all testable game state (no I/O dependencies)
type GameState struct {
	Board    *Board
	Current  *Tetromino
	Next     TetrominoType
	Lines    int
	GameOver bool
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	return &GameState{
		Board:    NewBoard(),
		GameOver: false,
	}
}

// TryMove attempts to move the current piece by the given offset
// Returns true if the move was successful
func (s *GameState) TryMove(dx, dy int) bool {
	if s.Current == nil {
		return false
	}
	moved := s.Current.Move(dx, dy)
	if s.Board.IsValidPosition(moved) {
		*s.Current = moved
		return true
	}
	return false
}

// TryRotate attempts to rotate the current piece clockwise with wall kicks
// Returns true if rotation was successful
func (s *GameState) TryRotate() bool {
	if s.Current == nil {
		return false
	}

	rotated := s.Current.RotateCW()
	kicks := s.Current.GetWallKicks(rotated.Rotation)

	for _, kick := range kicks {
		kicked := rotated.Move(kick.X, kick.Y)
		if s.Board.IsValidPosition(kicked) {
			*s.Current = kicked
			return true
		}
	}
	return false
}

// HardDrop instantly drops the piece to the bottom and locks it
func (s *GameState) HardDrop() {
	if s.Current == nil {
		return
	}

	// Move down until we can't
	for s.Board.IsValidPosition(s.Current.Move(0, 1)) {
		*s.Current = s.Current.Move(0, 1)
	}
}

// Tick advances the game by one gravity step
// Returns true if the piece was locked (triggers line clear check)
func (s *GameState) Tick() bool {
	if s.Current == nil || s.GameOver {
		return false
	}

	// Try to move down
	if s.TryMove(0, 1) {
		return false
	}

	// Can't move down - lock the piece
	s.Board.Lock(*s.Current)
	s.Current = nil
	return true
}

// LockAndClear locks the current piece and clears lines
// Returns the number of lines cleared
func (s *GameState) LockAndClear() int {
	if s.Current != nil {
		s.Board.Lock(*s.Current)
		s.Current = nil
	}

	linesCleared := s.Board.ClearLines()
	if linesCleared > 0 {
		s.Lines += linesCleared
	}

	return linesCleared
}

// SpawnPiece creates a new piece at the top of the board
// Returns false if the piece cannot be placed (game over)
func (s *GameState) SpawnPiece(nextType TetrominoType) bool {
	piece := NewTetromino(s.Next)
	s.Next = nextType

	if !s.Board.IsValidPosition(piece) {
		s.GameOver = true
		return false
	}

	s.Current = &piece
	return true
}

// CheckGameOver checks if the board is in a game over state
func (s *GameState) CheckGameOver() bool {
	if s.Board.IsGameOver() {
		s.GameOver = true
		return true
	}
	return false
}

// Game orchestrates gameplay with I/O dependencies
type Game struct {
	state      *GameState
	device     protocol.DeviceConnection
	renderer   *Renderer
	background []byte
	inputChan  chan rune
	running    bool
	randSource RandSource
}

// NewGame creates a new Tetris game
func NewGame(device protocol.DeviceConnection) *Game {
	return &Game{
		device:     device,
		renderer:   NewRenderer(device),
		inputChan:  make(chan rune, 10),
		running:    true,
		randSource: defaultRand{},
	}
}

// reset initializes the game state for a new game
func (g *Game) reset() {
	g.state = NewGameState()
	g.state.Next = TetrominoType(g.randSource.Intn(int(TetrominoCount)))
	g.background = GenerateGameBackground()
}

// spawnNextPiece spawns a new piece and generates the next piece type
func (g *Game) spawnNextPiece() bool {
	nextType := TetrominoType(g.randSource.Intn(int(TetrominoCount)))
	return g.state.SpawnPiece(nextType)
}

// handleInput processes keyboard input
func (g *Game) handleInput() {
	select {
	case key := <-g.inputChan:
		switch key {
		case 'a', 'A':
			g.state.TryMove(-1, 0)
		case 'd', 'D':
			g.state.TryMove(1, 0)
		case 'w', 'W':
			g.state.TryRotate()
		case 's', 'S':
			g.state.TryMove(0, 1) // Soft drop
		case ' ':
			g.state.HardDrop()
			g.state.LockAndClear()
			g.spawnNextPiece()
		case 'q', 'Q':
			g.running = false
		}
	default:
	}
}

// render draws the current game state to the display
func (g *Game) render() error {
	g.renderer.RenderState(g.state.Board, g.state.Current, g.background)
	return g.renderer.Flush()
}

// showImage displays a static image on the device
func (g *Game) showImage(rgbData []byte) error {
	if err := protocol.SetDrawMode(g.device, 1); err != nil {
		return err
	}
	if err := protocol.SendImage(g.device, rgbData); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}

// waitForKey blocks until a key is pressed
func (g *Game) waitForKey() rune {
	return <-g.inputChan
}

// startInputReader starts the keyboard input goroutine
func (g *Game) startInputReader() func() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Warning: could not set raw mode: %v\n", err)
		return func() {}
	}

	go func() {
		buf := make([]byte, 3)
		for g.running {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				// Handle escape sequences for arrow keys
				if n == 3 && buf[0] == 27 && buf[1] == 91 {
					switch buf[2] {
					case 65: // up arrow
						g.inputChan <- 'w'
					case 66: // down arrow
						g.inputChan <- 's'
					case 67: // right arrow
						g.inputChan <- 'd'
					case 68: // left arrow
						g.inputChan <- 'a'
					}
				} else {
					g.inputChan <- rune(buf[0])
				}
			}
		}
	}()

	return func() {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}
}

// runGame runs the main game loop
func (g *Game) runGame() {
	// Initialize renderer with background
	g.renderer.SetPrevBuffer(g.background)
	g.renderer.SetCurrBuffer(g.background)

	// Display initial background
	if err := g.showImage(g.background); err != nil {
		return
	}

	// Spawn first piece
	if !g.spawnNextPiece() {
		return
	}

	lastDrop := time.Now()
	lastRender := time.Now()

	for g.running && !g.state.GameOver {
		now := time.Now()

		// Handle input
		g.handleInput()

		// Gravity
		if now.Sub(lastDrop) >= DropInterval {
			locked := g.state.Tick()
			if locked {
				g.state.LockAndClear()
				if g.state.CheckGameOver() {
					break
				}
				if !g.spawnNextPiece() {
					break
				}
			}
			lastDrop = now
		}

		// Render
		if now.Sub(lastRender) >= RenderInterval {
			if err := g.render(); err != nil {
				fmt.Printf("Render error: %v\n", err)
			}
			lastRender = now
		}

		// Small sleep to avoid busy loop
		time.Sleep(10 * time.Millisecond)
	}
}

// Run starts the main game loop
func (g *Game) Run() error {
	fmt.Println("Starting Tetris!")
	fmt.Println("Controls: A/Left=Left, D/Right=Right, W/Up=Rotate, S/Down=Soft drop, Space=Hard drop, Q=Quit")

	cleanup := g.startInputReader()
	defer cleanup()

	for g.running {
		// Show cover image and wait for key to start
		if err := g.showImage(GenerateCoverImage()); err != nil {
			return err
		}
		fmt.Print("Press any key to start...")
		key := g.waitForKey()
		fmt.Println()
		if key == 'q' || key == 'Q' {
			break
		}

		// Reset and start game
		g.reset()
		g.runGame()

		if !g.running {
			break
		}

		// Show game over screen
		if err := g.showImage(GenerateGameOverImage()); err != nil {
			return err
		}
		fmt.Printf("Game Over! Lines: %d\n", g.state.Lines)
		fmt.Print("Press any key to restart (Q to quit)...")
		key = g.waitForKey()
		fmt.Println()
		if key == 'q' || key == 'Q' {
			break
		}
	}

	return nil
}

package tetris

import (
	"testing"
)

// mockRand provides deterministic random numbers for testing
type mockRand struct {
	values []int
	index  int
}

func (m *mockRand) Intn(n int) int {
	if m.index >= len(m.values) {
		return 0
	}
	v := m.values[m.index] % n
	m.index++
	return v
}

func TestGameStateTryMove(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*GameState)
		dx, dy   int
		expected bool
	}{
		{
			name: "move left when valid",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 5
				piece.Y = 5
				s.Current = &piece
			},
			dx:       -1,
			dy:       0,
			expected: true,
		},
		{
			name: "move left blocked by wall",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 0 // At left edge
				piece.Y = 5
				s.Current = &piece
			},
			dx:       -1,
			dy:       0,
			expected: false,
		},
		{
			name: "move left blocked by locked piece",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 3
				piece.Y = 5
				s.Current = &piece
				// Lock a piece to the left of where T would move
				s.Board.Cells[6][2].Occupied = true
			},
			dx:       -1,
			dy:       0,
			expected: false,
		},
		{
			name: "move down when valid",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 3
				piece.Y = 5
				s.Current = &piece
			},
			dx:       0,
			dy:       1,
			expected: true,
		},
		{
			name: "move down blocked by floor",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 3
				piece.Y = BoardHeight - 2 // At bottom
				s.Current = &piece
			},
			dx:       0,
			dy:       1,
			expected: false,
		},
		{
			name: "no current piece",
			setup: func(s *GameState) {
				s.Current = nil
			},
			dx:       -1,
			dy:       0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameState()
			tt.setup(state)

			result := state.TryMove(tt.dx, tt.dy)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGameStateTryRotate(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*GameState)
		expected bool
	}{
		{
			name: "rotate when space available",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 4
				piece.Y = 5
				s.Current = &piece
			},
			expected: true,
		},
		{
			name: "rotate with wall kick near left wall",
			setup: func(s *GameState) {
				piece := Tetromino{Type: TetrominoT, Rotation: Rotation90, X: 0, Y: 5}
				s.Current = &piece
			},
			expected: true, // Should wall kick right
		},
		{
			name: "O piece rotation always succeeds",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoO)
				piece.X = 4
				piece.Y = 5
				s.Current = &piece
			},
			expected: true,
		},
		{
			name: "no current piece",
			setup: func(s *GameState) {
				s.Current = nil
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameState()
			tt.setup(state)

			result := state.TryRotate()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGameStateHardDrop(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*GameState)
		expectedY    int
	}{
		{
			name: "drops to bottom on empty board",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 3
				piece.Y = 0
				s.Current = &piece
			},
			expectedY: BoardHeight - 2, // T piece extends 1 below anchor
		},
		{
			name: "drops onto locked pieces",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 3
				piece.Y = 0
				s.Current = &piece
				// Fill bottom row
				for x := 0; x < BoardWidth; x++ {
					s.Board.Cells[BoardHeight-1][x].Occupied = true
				}
			},
			expectedY: BoardHeight - 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameState()
			tt.setup(state)

			state.HardDrop()

			if state.Current.Y != tt.expectedY {
				t.Errorf("expected Y=%d, got Y=%d", tt.expectedY, state.Current.Y)
			}
		})
	}
}

func TestGameStateTick(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*GameState)
		expectedLocked bool
		expectedY      int
	}{
		{
			name: "piece moves down",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 3
				piece.Y = 5
				s.Current = &piece
			},
			expectedLocked: false,
			expectedY:      6,
		},
		{
			name: "piece locks at bottom",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 3
				piece.Y = BoardHeight - 2
				s.Current = &piece
			},
			expectedLocked: true,
			expectedY:      BoardHeight - 2, // Unchanged, but locked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameState()
			tt.setup(state)
			origY := state.Current.Y

			locked := state.Tick()

			if locked != tt.expectedLocked {
				t.Errorf("expected locked=%v, got %v", tt.expectedLocked, locked)
			}

			if tt.expectedLocked {
				if state.Current != nil {
					t.Error("current piece should be nil after locking")
				}
			} else {
				if state.Current.Y != origY+1 {
					t.Errorf("piece should have moved down, expected Y=%d, got Y=%d", origY+1, state.Current.Y)
				}
			}
		})
	}
}

func TestGameStateLockAndClear(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*GameState)
		expectedLines int
	}{
		{
			name: "no lines to clear",
			setup: func(s *GameState) {
				piece := NewTetromino(TetrominoT)
				piece.X = 0
				piece.Y = BoardHeight - 2
				s.Current = &piece
			},
			expectedLines: 0,
		},
		{
			name: "single line clear",
			setup: func(s *GameState) {
				// Fill bottom row except last column
				for x := 0; x < BoardWidth-1; x++ {
					s.Board.Cells[BoardHeight-1][x].Occupied = true
				}
				// I piece horizontal will fill the gap
				piece := Tetromino{Type: TetrominoI, Rotation: Rotation0, X: 6, Y: BoardHeight - 2}
				s.Current = &piece
			},
			expectedLines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameState()
			tt.setup(state)

			lines := state.LockAndClear()

			if lines != tt.expectedLines {
				t.Errorf("expected %d lines, got %d", tt.expectedLines, lines)
			}
			if state.Current != nil {
				t.Error("current piece should be nil after lock")
			}
		})
	}
}

func TestGameStateSpawnPiece(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*GameState)
		expectedSpawn  bool
		expectedGameOver bool
	}{
		{
			name: "spawn on empty board",
			setup: func(s *GameState) {
				s.Next = TetrominoT
			},
			expectedSpawn:    true,
			expectedGameOver: false,
		},
		{
			name: "spawn blocked by existing pieces",
			setup: func(s *GameState) {
				s.Next = TetrominoT
				// Block spawn position
				s.Board.Cells[0][4].Occupied = true
				s.Board.Cells[1][3].Occupied = true
				s.Board.Cells[1][4].Occupied = true
				s.Board.Cells[1][5].Occupied = true
			},
			expectedSpawn:    false,
			expectedGameOver: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGameState()
			tt.setup(state)

			spawned := state.SpawnPiece(TetrominoI)

			if spawned != tt.expectedSpawn {
				t.Errorf("expected spawn=%v, got %v", tt.expectedSpawn, spawned)
			}
			if state.GameOver != tt.expectedGameOver {
				t.Errorf("expected gameOver=%v, got %v", tt.expectedGameOver, state.GameOver)
			}
		})
	}
}

func TestNewGameState(t *testing.T) {
	state := NewGameState()

	if state.Board == nil {
		t.Error("board should not be nil")
	}
	if state.Current != nil {
		t.Error("current piece should be nil initially")
	}
	if state.Lines != 0 {
		t.Error("lines should be 0 initially")
	}
	if state.GameOver {
		t.Error("game should not be over initially")
	}
}

func TestGameStateCheckGameOver(t *testing.T) {
	state := NewGameState()

	// Empty board
	if state.CheckGameOver() {
		t.Error("empty board should not be game over")
	}

	// Occupied top row
	state.Board.Cells[0][5].Occupied = true
	if !state.CheckGameOver() {
		t.Error("occupied top row should be game over")
	}
	if !state.GameOver {
		t.Error("GameOver flag should be set")
	}
}

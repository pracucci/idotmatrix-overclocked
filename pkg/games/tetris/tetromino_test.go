package tetris

import (
	"testing"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

func TestTetrominoGetCells(t *testing.T) {
	tests := []struct {
		name     string
		tetromino Tetromino
		expected []graphic.Point
	}{
		{
			name:     "I piece at origin rotation 0",
			tetromino: Tetromino{Type: TetrominoI, Rotation: Rotation0, X: 0, Y: 0},
			expected: []graphic.Point{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
		},
		{
			name:     "I piece at origin rotation 90",
			tetromino: Tetromino{Type: TetrominoI, Rotation: Rotation90, X: 0, Y: 0},
			expected: []graphic.Point{{2, 0}, {2, 1}, {2, 2}, {2, 3}},
		},
		{
			name:     "O piece at origin",
			tetromino: Tetromino{Type: TetrominoO, Rotation: Rotation0, X: 0, Y: 0},
			expected: []graphic.Point{{1, 0}, {2, 0}, {1, 1}, {2, 1}},
		},
		{
			name:     "T piece at position (3,5)",
			tetromino: Tetromino{Type: TetrominoT, Rotation: Rotation0, X: 3, Y: 5},
			expected: []graphic.Point{{4, 5}, {3, 6}, {4, 6}, {5, 6}},
		},
		{
			name:     "L piece rotation 90",
			tetromino: Tetromino{Type: TetrominoL, Rotation: Rotation90, X: 0, Y: 0},
			expected: []graphic.Point{{1, 0}, {1, 1}, {1, 2}, {2, 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cells := tt.tetromino.GetCells()
			if len(cells) != 4 {
				t.Errorf("expected 4 cells, got %d", len(cells))
				return
			}
			// Check each expected cell is present
			for _, exp := range tt.expected {
				found := false
				for _, cell := range cells {
					if cell == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected cell %v not found in %v", exp, cells)
				}
			}
		})
	}
}

func TestTetrominoRotateCW(t *testing.T) {
	tests := []struct {
		name         string
		initial      Rotation
		expectedRot  Rotation
	}{
		{"0 to 90", Rotation0, Rotation90},
		{"90 to 180", Rotation90, Rotation180},
		{"180 to 270", Rotation180, Rotation270},
		{"270 to 0", Rotation270, Rotation0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := Tetromino{Type: TetrominoT, Rotation: tt.initial, X: 5, Y: 10}
			rotated := original.RotateCW()

			if rotated.Rotation != tt.expectedRot {
				t.Errorf("expected rotation %d, got %d", tt.expectedRot, rotated.Rotation)
			}
			// Verify position unchanged
			if rotated.X != original.X || rotated.Y != original.Y {
				t.Errorf("position changed: expected (%d,%d), got (%d,%d)",
					original.X, original.Y, rotated.X, rotated.Y)
			}
			// Verify type unchanged
			if rotated.Type != original.Type {
				t.Errorf("type changed")
			}
		})
	}
}

func TestTetrominoMove(t *testing.T) {
	original := Tetromino{Type: TetrominoJ, Rotation: Rotation0, X: 5, Y: 10}

	tests := []struct {
		name     string
		dx, dy   int
		expectedX, expectedY int
	}{
		{"move left", -1, 0, 4, 10},
		{"move right", 1, 0, 6, 10},
		{"move down", 0, 1, 5, 11},
		{"move up", 0, -1, 5, 9},
		{"move diagonal", 2, 3, 7, 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moved := original.Move(tt.dx, tt.dy)

			if moved.X != tt.expectedX || moved.Y != tt.expectedY {
				t.Errorf("expected (%d,%d), got (%d,%d)",
					tt.expectedX, tt.expectedY, moved.X, moved.Y)
			}
			// Verify original unchanged (immutability)
			if original.X != 5 || original.Y != 10 {
				t.Errorf("original was mutated")
			}
			// Verify rotation and type unchanged
			if moved.Rotation != original.Rotation || moved.Type != original.Type {
				t.Errorf("rotation or type changed")
			}
		})
	}
}

func TestOPieceRotationIsIdentical(t *testing.T) {
	// O piece should have the same cells in all rotations
	for rot := Rotation0; rot <= Rotation270; rot++ {
		tetro := Tetromino{Type: TetrominoO, Rotation: rot, X: 0, Y: 0}
		cells := tetro.GetCells()
		expected := []graphic.Point{{1, 0}, {2, 0}, {1, 1}, {2, 1}}

		for _, exp := range expected {
			found := false
			for _, cell := range cells {
				if cell == exp {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("O piece rotation %d: expected cell %v not found", rot, exp)
			}
		}
	}
}

func TestTetrominoColors(t *testing.T) {
	tests := []struct {
		pieceType TetrominoType
		expected  graphic.Color
	}{
		{TetrominoI, graphic.Color{0, 255, 255}},   // Cyan
		{TetrominoO, graphic.Color{255, 255, 0}},   // Yellow
		{TetrominoT, graphic.Color{160, 32, 240}},  // Purple
		{TetrominoS, graphic.Color{0, 255, 0}},     // Green
		{TetrominoZ, graphic.Color{255, 0, 0}},     // Red
		{TetrominoJ, graphic.Color{0, 0, 255}},     // Blue
		{TetrominoL, graphic.Color{255, 165, 0}},   // Orange
	}

	for _, tt := range tests {
		tetro := Tetromino{Type: tt.pieceType}
		color := tetro.GetColor()
		if color != tt.expected {
			t.Errorf("piece type %d: expected color %v, got %v", tt.pieceType, tt.expected, color)
		}
	}
}

func TestNewTetromino(t *testing.T) {
	for pieceType := TetrominoI; pieceType < TetrominoCount; pieceType++ {
		tetro := NewTetromino(pieceType)

		if tetro.Type != pieceType {
			t.Errorf("expected type %d, got %d", pieceType, tetro.Type)
		}
		if tetro.Rotation != Rotation0 {
			t.Errorf("expected rotation 0, got %d", tetro.Rotation)
		}
		if tetro.X != 3 {
			t.Errorf("expected X=3, got %d", tetro.X)
		}
		if tetro.Y != 0 {
			t.Errorf("expected Y=0, got %d", tetro.Y)
		}
	}
}

func TestEachPieceHasFourCells(t *testing.T) {
	for pieceType := TetrominoI; pieceType < TetrominoCount; pieceType++ {
		for rot := Rotation0; rot <= Rotation270; rot++ {
			tetro := Tetromino{Type: pieceType, Rotation: rot, X: 0, Y: 0}
			cells := tetro.GetCells()
			if len(cells) != 4 {
				t.Errorf("piece type %d rotation %d: expected 4 cells, got %d", pieceType, rot, len(cells))
			}
		}
	}
}

func TestWallKicksExist(t *testing.T) {
	// Test that wall kicks return non-empty slices
	for pieceType := TetrominoI; pieceType < TetrominoCount; pieceType++ {
		tetro := Tetromino{Type: pieceType, Rotation: Rotation0}
		rotated := tetro.RotateCW()
		kicks := tetro.GetWallKicks(rotated.Rotation)

		if len(kicks) == 0 {
			t.Errorf("piece type %d: wall kicks should not be empty", pieceType)
		}
		// First kick should always be (0,0)
		if kicks[0] != (graphic.Point{0, 0}) {
			t.Errorf("piece type %d: first wall kick should be (0,0), got %v", pieceType, kicks[0])
		}
	}
}

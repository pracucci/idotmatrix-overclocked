package tetris

import (
	"testing"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

func TestBoardIsValidPosition(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Board)
		tetromino Tetromino
		expected bool
	}{
		{
			name:     "piece within bounds valid",
			setup:    func(b *Board) {},
			tetromino: Tetromino{Type: TetrominoT, Rotation: Rotation0, X: 3, Y: 5},
			expected: true,
		},
		{
			name:     "piece left of board invalid",
			setup:    func(b *Board) {},
			tetromino: Tetromino{Type: TetrominoT, Rotation: Rotation0, X: -1, Y: 5},
			expected: false,
		},
		{
			name:     "piece right of board invalid",
			setup:    func(b *Board) {},
			tetromino: Tetromino{Type: TetrominoI, Rotation: Rotation0, X: 8, Y: 5}, // I piece is 4 wide
			expected: false,
		},
		{
			name:     "piece below board invalid",
			setup:    func(b *Board) {},
			tetromino: Tetromino{Type: TetrominoT, Rotation: Rotation0, X: 3, Y: 19}, // T at row 19, extends to 20
			expected: false,
		},
		{
			name: "piece overlapping locked cell invalid",
			setup: func(b *Board) {
				b.Cells[6][4].Occupied = true // Block where T piece would go
			},
			tetromino: Tetromino{Type: TetrominoT, Rotation: Rotation0, X: 3, Y: 5}, // T at (3,5) has cell at (4,6)
			expected: false,
		},
		{
			name: "piece in empty space with locked cells nearby valid",
			setup: func(b *Board) {
				b.Cells[7][4].Occupied = true // Below where piece would be
				b.Cells[6][2].Occupied = true // Left of where piece would be
			},
			tetromino: Tetromino{Type: TetrominoT, Rotation: Rotation0, X: 3, Y: 5},
			expected: true,
		},
		{
			name:     "piece at spawn position valid on empty board",
			setup:    func(b *Board) {},
			tetromino: NewTetromino(TetrominoI),
			expected: true,
		},
		{
			name:     "piece partially above board valid",
			setup:    func(b *Board) {},
			tetromino: Tetromino{Type: TetrominoI, Rotation: Rotation90, X: 0, Y: -2}, // Vertical I piece
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setup(board)

			result := board.IsValidPosition(tt.tetromino)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBoardLock(t *testing.T) {
	board := NewBoard()
	tetro := Tetromino{Type: TetrominoT, Rotation: Rotation0, X: 3, Y: 5}

	board.Lock(tetro)

	// T piece at (3,5) rotation 0 occupies: (4,5), (3,6), (4,6), (5,6)
	expectedCells := []graphic.Point{{4, 5}, {3, 6}, {4, 6}, {5, 6}}
	expectedColor := tetro.GetColor()

	for _, p := range expectedCells {
		cell := board.GetCell(p.X, p.Y)
		if !cell.Occupied {
			t.Errorf("cell at (%d,%d) should be occupied", p.X, p.Y)
		}
		if cell.Color != expectedColor {
			t.Errorf("cell at (%d,%d) has wrong color: expected %v, got %v",
				p.X, p.Y, expectedColor, cell.Color)
		}
	}

	// Check that adjacent cells are not occupied
	notOccupied := []graphic.Point{{2, 5}, {2, 6}, {6, 6}, {4, 7}}
	for _, p := range notOccupied {
		cell := board.GetCell(p.X, p.Y)
		if cell.Occupied {
			t.Errorf("cell at (%d,%d) should not be occupied", p.X, p.Y)
		}
	}
}

func TestBoardClearLines(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*Board)
		expectedLines int
		checkAfter    func(*testing.T, *Board)
	}{
		{
			name: "no complete lines returns 0",
			setup: func(b *Board) {
				// Partial line at bottom
				for x := 0; x < BoardWidth-1; x++ {
					b.Cells[BoardHeight-1][x].Occupied = true
				}
			},
			expectedLines: 0,
			checkAfter:    func(t *testing.T, b *Board) {},
		},
		{
			name: "one complete line returns 1",
			setup: func(b *Board) {
				// Complete line at bottom
				for x := 0; x < BoardWidth; x++ {
					b.Cells[BoardHeight-1][x].Occupied = true
					b.Cells[BoardHeight-1][x].Color = graphic.Color{255, 0, 0}
				}
				// Partial line above
				b.Cells[BoardHeight-2][0].Occupied = true
				b.Cells[BoardHeight-2][0].Color = graphic.Color{0, 255, 0}
			},
			expectedLines: 1,
			checkAfter: func(t *testing.T, b *Board) {
				// The partial line should have dropped to bottom
				if !b.Cells[BoardHeight-1][0].Occupied {
					t.Error("cell should have dropped to bottom row")
				}
				if b.Cells[BoardHeight-1][0].Color != (graphic.Color{0, 255, 0}) {
					t.Error("dropped cell should have green color")
				}
				// Rest of bottom row should be empty now
				for x := 1; x < BoardWidth; x++ {
					if b.Cells[BoardHeight-1][x].Occupied {
						t.Errorf("cell at (%d, %d) should be empty after line clear", x, BoardHeight-1)
					}
				}
			},
		},
		{
			name: "two lines cleared at once returns 2",
			setup: func(b *Board) {
				// Two complete lines at bottom
				for y := BoardHeight - 2; y < BoardHeight; y++ {
					for x := 0; x < BoardWidth; x++ {
						b.Cells[y][x].Occupied = true
					}
				}
			},
			expectedLines: 2,
			checkAfter: func(t *testing.T, b *Board) {
				// Both rows should be empty
				for y := BoardHeight - 2; y < BoardHeight; y++ {
					for x := 0; x < BoardWidth; x++ {
						if b.Cells[y][x].Occupied {
							t.Errorf("cell at (%d,%d) should be empty", x, y)
						}
					}
				}
			},
		},
		{
			name: "clearing middle line drops above",
			setup: func(b *Board) {
				// Complete line in middle
				for x := 0; x < BoardWidth; x++ {
					b.Cells[10][x].Occupied = true
				}
				// Single cell above
				b.Cells[9][5].Occupied = true
				b.Cells[9][5].Color = graphic.Color{0, 0, 255}
			},
			expectedLines: 1,
			checkAfter: func(t *testing.T, b *Board) {
				// Cell should have dropped
				if !b.Cells[10][5].Occupied {
					t.Error("cell should have dropped to row 10")
				}
				if b.Cells[9][5].Occupied {
					t.Error("row 9 should be empty after drop")
				}
			},
		},
		{
			name: "tetris (4 lines) returns 4",
			setup: func(b *Board) {
				for y := BoardHeight - 4; y < BoardHeight; y++ {
					for x := 0; x < BoardWidth; x++ {
						b.Cells[y][x].Occupied = true
					}
				}
			},
			expectedLines: 4,
			checkAfter:    func(t *testing.T, b *Board) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setup(board)

			lines := board.ClearLines()
			if lines != tt.expectedLines {
				t.Errorf("expected %d lines cleared, got %d", tt.expectedLines, lines)
			}

			tt.checkAfter(t, board)
		})
	}
}

func TestBoardIsGameOver(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Board)
		expected bool
	}{
		{
			name:     "empty board not game over",
			setup:    func(b *Board) {},
			expected: false,
		},
		{
			name: "pieces at bottom not game over",
			setup: func(b *Board) {
				for x := 0; x < BoardWidth; x++ {
					b.Cells[BoardHeight-1][x].Occupied = true
				}
			},
			expected: false,
		},
		{
			name: "cell in row 0 is game over",
			setup: func(b *Board) {
				b.Cells[0][5].Occupied = true
			},
			expected: true,
		},
		{
			name: "cell in row 1 is game over",
			setup: func(b *Board) {
				b.Cells[1][3].Occupied = true
			},
			expected: true,
		},
		{
			name: "cell in row 2 not game over",
			setup: func(b *Board) {
				b.Cells[2][5].Occupied = true
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard()
			tt.setup(board)

			result := board.IsGameOver()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

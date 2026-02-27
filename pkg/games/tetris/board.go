package tetris

import "github.com/pracucci/idotmatrix-overclocked/pkg/graphic"

// Board dimensions
const (
	BoardWidth  = 10 // Standard Tetris board width
	BoardHeight = 20 // Standard Tetris board height
)

// Cell represents a single cell on the game board
type Cell struct {
	Occupied bool
	Color    graphic.Color
}

// Board represents the game board state
type Board struct {
	Cells [BoardHeight][BoardWidth]Cell
}

// NewBoard creates a new empty board
func NewBoard() *Board {
	return &Board{}
}

// IsValidPosition checks if a tetromino can be placed at its current position
// Returns true if all cells are within bounds and not overlapping locked pieces
func (b *Board) IsValidPosition(t Tetromino) bool {
	cells := t.GetCells()
	for _, cell := range cells {
		// Check horizontal bounds
		if cell.X < 0 || cell.X >= BoardWidth {
			return false
		}
		// Check vertical bounds (allow negative Y for pieces spawning above)
		if cell.Y >= BoardHeight {
			return false
		}
		// Check collision with locked pieces (only if cell is on the board)
		if cell.Y >= 0 && b.Cells[cell.Y][cell.X].Occupied {
			return false
		}
	}
	return true
}

// Lock places a tetromino on the board, marking its cells as occupied
func (b *Board) Lock(t Tetromino) {
	cells := t.GetCells()
	color := t.GetColor()
	for _, cell := range cells {
		if cell.Y >= 0 && cell.Y < BoardHeight && cell.X >= 0 && cell.X < BoardWidth {
			b.Cells[cell.Y][cell.X] = Cell{
				Occupied: true,
				Color:    color,
			}
		}
	}
}

// ClearLines removes all complete lines and returns the number cleared
// Lines above the cleared lines drop down
func (b *Board) ClearLines() int {
	linesCleared := 0

	// Check each row from bottom to top
	y := BoardHeight - 1
	for y >= 0 {
		if b.isLineComplete(y) {
			b.clearLine(y)
			linesCleared++
			// Don't decrement y - check this row again as lines dropped
		} else {
			y--
		}
	}

	return linesCleared
}

// isLineComplete checks if a row is completely filled
func (b *Board) isLineComplete(y int) bool {
	for x := 0; x < BoardWidth; x++ {
		if !b.Cells[y][x].Occupied {
			return false
		}
	}
	return true
}

// clearLine removes a line and drops all lines above it
func (b *Board) clearLine(clearedY int) {
	// Move all lines above down by one
	for y := clearedY; y > 0; y-- {
		for x := 0; x < BoardWidth; x++ {
			b.Cells[y][x] = b.Cells[y-1][x]
		}
	}
	// Clear the top line
	for x := 0; x < BoardWidth; x++ {
		b.Cells[0][x] = Cell{}
	}
}

// IsGameOver checks if the game is over (pieces stacked to the top)
// This checks if any cells in the top two rows are occupied
func (b *Board) IsGameOver() bool {
	// Check rows 0 and 1 (spawn area)
	for y := 0; y < 2; y++ {
		for x := 0; x < BoardWidth; x++ {
			if b.Cells[y][x].Occupied {
				return true
			}
		}
	}
	return false
}

// GetCell returns the cell at the given position
// Returns an empty cell if position is out of bounds
func (b *Board) GetCell(x, y int) Cell {
	if x < 0 || x >= BoardWidth || y < 0 || y >= BoardHeight {
		return Cell{}
	}
	return b.Cells[y][x]
}

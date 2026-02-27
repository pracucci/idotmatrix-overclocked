package tetris

import "github.com/pracucci/idotmatrix-overclocked/pkg/graphic"

// TetrominoType represents the 7 standard tetromino pieces
type TetrominoType int

const (
	TetrominoI TetrominoType = iota
	TetrominoO
	TetrominoT
	TetrominoS
	TetrominoZ
	TetrominoJ
	TetrominoL
	TetrominoCount // Used for iteration/random selection
)

// Rotation represents the rotation state of a tetromino
type Rotation int

const (
	Rotation0   Rotation = iota // 0 degrees (spawn state)
	Rotation90                  // 90 degrees clockwise
	Rotation180                 // 180 degrees
	Rotation270                 // 270 degrees clockwise
)

// Tetromino represents a tetris piece with position and rotation
// Tetromino is immutable - all methods return new instances
type Tetromino struct {
	Type     TetrominoType
	Rotation Rotation
	X, Y     int // Position in board coordinates (top-left of bounding box)
}

// tetrominoShapes defines the shape of each piece at each rotation
// Each shape is a 4x4 grid stored as offsets from position
// Format: [type][rotation][cellIndex] = graphic.Point offset
var tetrominoShapes = [TetrominoCount][4][]graphic.Point{
	// I piece - cyan
	TetrominoI: {
		{{0, 1}, {1, 1}, {2, 1}, {3, 1}}, // 0
		{{2, 0}, {2, 1}, {2, 2}, {2, 3}}, // 90
		{{0, 2}, {1, 2}, {2, 2}, {3, 2}}, // 180
		{{1, 0}, {1, 1}, {1, 2}, {1, 3}}, // 270
	},
	// O piece - yellow
	TetrominoO: {
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}}, // 0
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}}, // 90 (same)
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}}, // 180 (same)
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}}, // 270 (same)
	},
	// T piece - purple
	TetrominoT: {
		{{1, 0}, {0, 1}, {1, 1}, {2, 1}}, // 0
		{{1, 0}, {1, 1}, {2, 1}, {1, 2}}, // 90
		{{0, 1}, {1, 1}, {2, 1}, {1, 2}}, // 180
		{{1, 0}, {0, 1}, {1, 1}, {1, 2}}, // 270
	},
	// S piece - green
	TetrominoS: {
		{{1, 0}, {2, 0}, {0, 1}, {1, 1}}, // 0
		{{1, 0}, {1, 1}, {2, 1}, {2, 2}}, // 90
		{{1, 1}, {2, 1}, {0, 2}, {1, 2}}, // 180
		{{0, 0}, {0, 1}, {1, 1}, {1, 2}}, // 270
	},
	// Z piece - red
	TetrominoZ: {
		{{0, 0}, {1, 0}, {1, 1}, {2, 1}}, // 0
		{{2, 0}, {1, 1}, {2, 1}, {1, 2}}, // 90
		{{0, 1}, {1, 1}, {1, 2}, {2, 2}}, // 180
		{{1, 0}, {0, 1}, {1, 1}, {0, 2}}, // 270
	},
	// J piece - blue
	TetrominoJ: {
		{{0, 0}, {0, 1}, {1, 1}, {2, 1}}, // 0
		{{1, 0}, {2, 0}, {1, 1}, {1, 2}}, // 90
		{{0, 1}, {1, 1}, {2, 1}, {2, 2}}, // 180
		{{1, 0}, {1, 1}, {0, 2}, {1, 2}}, // 270
	},
	// L piece - orange
	TetrominoL: {
		{{2, 0}, {0, 1}, {1, 1}, {2, 1}}, // 0
		{{1, 0}, {1, 1}, {1, 2}, {2, 2}}, // 90
		{{0, 1}, {1, 1}, {2, 1}, {0, 2}}, // 180
		{{0, 0}, {1, 0}, {1, 1}, {1, 2}}, // 270
	},
}

// tetrominoColors defines the color for each piece type
var tetrominoColors = [TetrominoCount]graphic.Color{
	TetrominoI: {0, 255, 255},   // Cyan
	TetrominoO: {255, 255, 0},   // Yellow
	TetrominoT: {160, 32, 240},  // Purple
	TetrominoS: {0, 255, 0},     // Green
	TetrominoZ: {255, 0, 0},     // Red
	TetrominoJ: {0, 0, 255},     // Blue
	TetrominoL: {255, 165, 0},   // Orange
}

// NewTetromino creates a new tetromino at the spawn position
func NewTetromino(t TetrominoType) Tetromino {
	return Tetromino{
		Type:     t,
		Rotation: Rotation0,
		X:        3, // Center horizontally (board width 10, piece width ~4)
		Y:        0, // Top of board
	}
}

// GetCells returns the absolute board positions occupied by this tetromino
func (t Tetromino) GetCells() []graphic.Point {
	offsets := tetrominoShapes[t.Type][t.Rotation]
	cells := make([]graphic.Point, len(offsets))
	for i, offset := range offsets {
		cells[i] = graphic.Point{
			X: t.X + offset.X,
			Y: t.Y + offset.Y,
		}
	}
	return cells
}

// GetColor returns the color for this tetromino type
func (t Tetromino) GetColor() graphic.Color {
	return tetrominoColors[t.Type]
}

// RotateCW returns a new tetromino rotated 90 degrees clockwise
func (t Tetromino) RotateCW() Tetromino {
	return Tetromino{
		Type:     t.Type,
		Rotation: (t.Rotation + 1) % 4,
		X:        t.X,
		Y:        t.Y,
	}
}

// Move returns a new tetromino moved by the given offset
func (t Tetromino) Move(dx, dy int) Tetromino {
	return Tetromino{
		Type:     t.Type,
		Rotation: t.Rotation,
		X:        t.X + dx,
		Y:        t.Y + dy,
	}
}

// GetWallKicks returns the wall kick offsets to try when rotating
// Uses SRS (Super Rotation System) wall kick data
func (t Tetromino) GetWallKicks(newRotation Rotation) []graphic.Point {
	if t.Type == TetrominoI {
		return getIWallKicks(t.Rotation, newRotation)
	}
	if t.Type == TetrominoO {
		return []graphic.Point{{0, 0}} // O piece doesn't need wall kicks
	}
	return getJLSTZWallKicks(t.Rotation, newRotation)
}

// SRS wall kick data for J, L, S, T, Z pieces
// Format: [fromRotation][toRotation] = []graphic.Point offsets to try
func getJLSTZWallKicks(from, to Rotation) []graphic.Point {
	// Test 0 is always (0,0)
	// These are the SRS offsets for JLSTZ pieces
	kickTable := map[[2]Rotation][]graphic.Point{
		{Rotation0, Rotation90}:   {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},
		{Rotation90, Rotation0}:   {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},
		{Rotation90, Rotation180}: {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},
		{Rotation180, Rotation90}: {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},
		{Rotation180, Rotation270}: {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},
		{Rotation270, Rotation180}: {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}},
		{Rotation270, Rotation0}: {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}},
		{Rotation0, Rotation270}: {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},
	}
	key := [2]Rotation{from, to}
	if kicks, ok := kickTable[key]; ok {
		return kicks
	}
	return []graphic.Point{{0, 0}}
}

// SRS wall kick data for I piece
func getIWallKicks(from, to Rotation) []graphic.Point {
	kickTable := map[[2]Rotation][]graphic.Point{
		{Rotation0, Rotation90}:   {{0, 0}, {-2, 0}, {1, 0}, {-2, 1}, {1, -2}},
		{Rotation90, Rotation0}:   {{0, 0}, {2, 0}, {-1, 0}, {2, -1}, {-1, 2}},
		{Rotation90, Rotation180}: {{0, 0}, {-1, 0}, {2, 0}, {-1, -2}, {2, 1}},
		{Rotation180, Rotation90}: {{0, 0}, {1, 0}, {-2, 0}, {1, 2}, {-2, -1}},
		{Rotation180, Rotation270}: {{0, 0}, {2, 0}, {-1, 0}, {2, -1}, {-1, 2}},
		{Rotation270, Rotation180}: {{0, 0}, {-2, 0}, {1, 0}, {-2, 1}, {1, -2}},
		{Rotation270, Rotation0}: {{0, 0}, {1, 0}, {-2, 0}, {1, 2}, {-2, -1}},
		{Rotation0, Rotation270}: {{0, 0}, {-1, 0}, {2, 0}, {-1, -2}, {2, 1}},
	}
	key := [2]Rotation{from, to}
	if kicks, ok := kickTable[key]; ok {
		return kicks
	}
	return []graphic.Point{{0, 0}}
}

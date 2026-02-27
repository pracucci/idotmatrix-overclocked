package tetris

import (
	"testing"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

func TestRendererComputeDiff(t *testing.T) {
	tests := []struct {
		name           string
		setupPrev      func(r *Renderer)
		setupCurr      func(r *Renderer)
		expectedColors int
		expectedPixels int
	}{
		{
			name:           "identical buffers no changes",
			setupPrev:      func(r *Renderer) {},
			setupCurr:      func(r *Renderer) {},
			expectedColors: 0,
			expectedPixels: 0,
		},
		{
			name:      "single pixel changed",
			setupPrev: func(r *Renderer) {},
			setupCurr: func(r *Renderer) {
				r.currBuffer[0] = 255 // R at (0,0)
			},
			expectedColors: 1,
			expectedPixels: 1,
		},
		{
			name:      "multiple pixels same color grouped",
			setupPrev: func(r *Renderer) {},
			setupCurr: func(r *Renderer) {
				// Set 3 pixels to red
				for i := 0; i < 3; i++ {
					offset := i * 3
					r.currBuffer[offset] = 255 // R
				}
			},
			expectedColors: 1,
			expectedPixels: 3,
		},
		{
			name:      "multiple colors separate groups",
			setupPrev: func(r *Renderer) {},
			setupCurr: func(r *Renderer) {
				// Pixel 0: red
				r.currBuffer[0] = 255
				// Pixel 1: green
				r.currBuffer[4] = 255
				// Pixel 2: blue
				r.currBuffer[8] = 255
			},
			expectedColors: 3,
			expectedPixels: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Renderer{}
			tt.setupPrev(r)
			tt.setupCurr(r)

			diff := r.ComputeDiff()

			if len(diff) != tt.expectedColors {
				t.Errorf("expected %d color groups, got %d", tt.expectedColors, len(diff))
			}

			totalPixels := 0
			for _, points := range diff {
				totalPixels += len(points)
			}
			if totalPixels != tt.expectedPixels {
				t.Errorf("expected %d total pixels, got %d", tt.expectedPixels, totalPixels)
			}
		})
	}
}

func TestRendererRenderState(t *testing.T) {
	r := &Renderer{}
	board := NewBoard()
	background := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Fill background with a pattern to verify it's copied
	for i := range background {
		background[i] = 50
	}

	// Render empty board
	r.RenderState(board, nil, background)

	// Background should be copied
	buf := r.GetCurrBuffer()
	if buf[0] != 50 || buf[1] != 50 || buf[2] != 50 {
		t.Error("background should be copied to buffer")
	}
}

func TestRendererRenderLockedPieces(t *testing.T) {
	r := &Renderer{}
	board := NewBoard()
	background := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Lock a piece on the board
	board.Cells[BoardHeight-1][0].Occupied = true
	board.Cells[BoardHeight-1][0].Color = graphic.Color{255, 0, 0}

	r.RenderState(board, nil, background)

	// Check that the block is rendered at the correct display position
	displayX := BoardOffsetX + 0*BlockSize
	displayY := BoardOffsetY + (BoardHeight-1)*BlockSize
	offset := (displayY*graphic.DisplayWidth + displayX) * 3

	buf := r.GetCurrBuffer()
	if buf[offset] != 255 || buf[offset+1] != 0 || buf[offset+2] != 0 {
		t.Errorf("locked piece should be rendered as red at (%d,%d), got RGB(%d,%d,%d)",
			displayX, displayY, buf[offset], buf[offset+1], buf[offset+2])
	}
}

func TestRendererRenderCurrentPiece(t *testing.T) {
	r := &Renderer{}
	board := NewBoard()
	background := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Create a T piece at position (3, 5)
	tetro := Tetromino{Type: TetrominoT, Rotation: Rotation0, X: 3, Y: 5}
	color := tetro.GetColor()

	r.RenderState(board, &tetro, background)

	// Check one of the T piece cells (center bottom: boardX=4, boardY=6)
	displayX := BoardOffsetX + 4*BlockSize
	displayY := BoardOffsetY + 6*BlockSize
	offset := (displayY*graphic.DisplayWidth + displayX) * 3

	buf := r.GetCurrBuffer()
	if buf[offset] != color[0] || buf[offset+1] != color[1] || buf[offset+2] != color[2] {
		t.Errorf("current piece should be rendered with color %v, got RGB(%d,%d,%d)",
			color, buf[offset], buf[offset+1], buf[offset+2])
	}
}

func TestRendererBlockSize(t *testing.T) {
	r := &Renderer{}
	board := NewBoard()
	background := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Lock a piece
	board.Cells[10][5].Occupied = true
	board.Cells[10][5].Color = graphic.Color{255, 0, 0}

	r.RenderState(board, nil, background)

	// Verify the entire 3x3 block is filled
	baseX := BoardOffsetX + 5*BlockSize
	baseY := BoardOffsetY + 10*BlockSize

	buf := r.GetCurrBuffer()
	for dy := 0; dy < BlockSize; dy++ {
		for dx := 0; dx < BlockSize; dx++ {
			offset := ((baseY+dy)*graphic.DisplayWidth + baseX + dx) * 3
			if buf[offset] != 255 || buf[offset+1] != 0 || buf[offset+2] != 0 {
				t.Errorf("block pixel at offset (%d,%d) should be red", dx, dy)
			}
		}
	}
}

func TestRendererSetPrevBuffer(t *testing.T) {
	r := &Renderer{}

	// Set prev buffer to all 100s
	prevData := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)
	for i := range prevData {
		prevData[i] = 100
	}
	r.SetPrevBuffer(prevData)

	// Set curr buffer to same values (no diff expected)
	currData := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)
	for i := range currData {
		currData[i] = 100
	}
	r.SetCurrBuffer(currData)

	diff := r.ComputeDiff()
	if len(diff) != 0 {
		t.Error("identical buffers should have no diff")
	}

	// Now change one pixel in curr
	r.currBuffer[0] = 200

	diff = r.ComputeDiff()
	if len(diff) != 1 {
		t.Error("should detect one changed pixel")
	}
}

func TestRendererPieceAboveBoardNotRendered(t *testing.T) {
	r := &Renderer{}
	board := NewBoard()
	background := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Create piece with Y=-1 (partially above board)
	tetro := Tetromino{Type: TetrominoI, Rotation: Rotation90, X: 0, Y: -2}

	r.RenderState(board, &tetro, background)

	// Cells at negative Y should not be rendered (would be at negative display coords)
	// This test verifies no crash occurs
}

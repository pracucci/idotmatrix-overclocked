package tetris

import (
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/text"
)

// drawTetromino draws a small tetromino shape for decoration
func drawTetromino(img []byte, pieceType TetrominoType, x, y int, scale int) {
	colors := []graphic.Color{graphic.Cyan, graphic.Yellow, graphic.Violet, graphic.Green, graphic.Red, graphic.Blue, graphic.Orange}
	color := colors[pieceType]

	offsets := tetrominoShapes[pieceType][Rotation0]
	for _, off := range offsets {
		for dy := 0; dy < scale; dy++ {
			for dx := 0; dx < scale; dx++ {
				graphic.SetPixel(img, x+off.X*scale+dx, y+off.Y*scale+dy, color)
			}
		}
	}
}

// GenerateCoverImage creates the title screen with "TETRIS" text and decorative pieces
func GenerateCoverImage() []byte {
	img := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Dark purple gradient background
	for y := 0; y < graphic.DisplayWidth; y++ {
		for x := 0; x < graphic.DisplayWidth; x++ {
			intensity := uint8(10 + y/8)
			graphic.SetPixel(img, x, y, graphic.Color{intensity / 4, 0, intensity / 2})
		}
	}

	// Draw "TETRIS" title (6 chars * 6 pixels = 36 pixels wide, center at (64-36)/2 = 14)
	// Draw shadow first
	text.DrawText(img, "TETRIS", 15, 9, graphic.DarkWhite)
	// Draw main text in cyan
	text.DrawText(img, "TETRIS", 14, 8, graphic.Cyan)

	// Draw decorative tetrominoes
	drawTetromino(img, TetrominoI, 2, 24, 2)
	drawTetromino(img, TetrominoO, 50, 20, 2)
	drawTetromino(img, TetrominoT, 6, 42, 2)
	drawTetromino(img, TetrominoS, 42, 38, 2)
	drawTetromino(img, TetrominoZ, 24, 50, 2)
	drawTetromino(img, TetrominoL, 48, 52, 2)
	drawTetromino(img, TetrominoJ, 4, 54, 2)

	// Draw a small decorative border at bottom
	for x := 10; x < 54; x++ {
		if x%3 != 0 {
			graphic.SetPixel(img, x, 60, graphic.DimGray)
		}
	}

	return img
}

// GenerateGameOverImage creates the game over screen
func GenerateGameOverImage() []byte {
	img := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Dark red tinted background
	for y := 0; y < graphic.DisplayWidth; y++ {
		for x := 0; x < graphic.DisplayWidth; x++ {
			intensity := uint8(20 + y/4)
			graphic.SetPixel(img, x, y, graphic.Color{intensity / 2, 0, 0})
		}
	}

	// Draw "GAME" and "OVER" text centered
	// "GAME" is 4 chars * 6 = 24 pixels, center at (64-24)/2 = 20
	text.DrawText(img, "GAME", 21, 25, graphic.DarkRed)
	text.DrawText(img, "GAME", 20, 24, graphic.Red)
	text.DrawText(img, "OVER", 21, 35, graphic.DarkRed)
	text.DrawText(img, "OVER", 20, 34, graphic.Red)

	return img
}

// GenerateGameBackground creates the background for gameplay
func GenerateGameBackground() []byte {
	img := make([]byte, graphic.DisplayWidth*graphic.DisplayWidth*3)

	// Fill with dark background
	for y := 0; y < graphic.DisplayWidth; y++ {
		for x := 0; x < graphic.DisplayWidth; x++ {
			graphic.SetPixel(img, x, y, graphic.Color{10, 10, 15})
		}
	}

	// Draw game board border
	boardLeft := BoardOffsetX - 1
	boardRight := BoardOffsetX + BoardWidth*BlockSize
	boardTop := BoardOffsetY - 1
	boardBottom := BoardOffsetY + BoardHeight*BlockSize

	// Left border
	for y := boardTop; y <= boardBottom; y++ {
		graphic.SetPixel(img, boardLeft, y, graphic.DimGray)
	}
	// Right border
	for y := boardTop; y <= boardBottom; y++ {
		graphic.SetPixel(img, boardRight, y, graphic.DimGray)
	}
	// Bottom border
	for x := boardLeft; x <= boardRight; x++ {
		graphic.SetPixel(img, x, boardBottom, graphic.DimGray)
	}
	// Top border (partial, for aesthetics)
	for x := boardLeft; x <= boardRight; x++ {
		if x%4 == 0 {
			graphic.SetPixel(img, x, boardTop, graphic.DarkWhite)
		}
	}

	// Draw subtle grid inside the board
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			displayX := BoardOffsetX + x*BlockSize
			displayY := BoardOffsetY + y*BlockSize
			// Draw subtle corner dots for grid
			graphic.SetPixel(img, displayX, displayY, graphic.Color{20, 20, 25})
		}
	}

	return img
}

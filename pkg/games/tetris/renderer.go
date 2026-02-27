package tetris

import (
	"time"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
)

// Rendering constants
const (
	BlockSize    = 3  // Each tetris cell = 3x3 display pixels
	BoardOffsetX = 17 // X offset to center board: (64 - 10*3) / 2
	BoardOffsetY = 2  // Y offset to center board: (64 - 20*3) / 2
)

// Renderer handles diff-based rendering to the device
type Renderer struct {
	device     protocol.DeviceConnection
	prevBuffer [graphic.DisplayWidth * graphic.DisplayWidth * 3]byte
	currBuffer [graphic.DisplayWidth * graphic.DisplayWidth * 3]byte
}

// NewRenderer creates a new renderer
func NewRenderer(device protocol.DeviceConnection) *Renderer {
	return &Renderer{
		device: device,
	}
}

// RenderState converts game state to the pixel buffer
// This is a pure function that updates currBuffer without I/O
func (r *Renderer) RenderState(board *Board, current *Tetromino, background []byte) {
	// Start with background
	copy(r.currBuffer[:], background)

	// Draw locked pieces on the board
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			cell := board.GetCell(x, y)
			if cell.Occupied {
				r.drawBlock(x, y, cell.Color)
			}
		}
	}

	// Draw current piece
	if current != nil {
		cells := current.GetCells()
		color := current.GetColor()
		for _, cell := range cells {
			if cell.Y >= 0 { // Only draw cells that are on the visible board
				r.drawBlock(cell.X, cell.Y, color)
			}
		}
	}
}

// drawBlock draws a single tetris block (3x3 pixels) on the buffer
func (r *Renderer) drawBlock(boardX, boardY int, color graphic.Color) {
	// Convert board coordinates to display coordinates
	displayX := BoardOffsetX + boardX*BlockSize
	displayY := BoardOffsetY + boardY*BlockSize

	// Draw 3x3 block
	for dy := 0; dy < BlockSize; dy++ {
		for dx := 0; dx < BlockSize; dx++ {
			px := displayX + dx
			py := displayY + dy
			if px >= 0 && px < graphic.DisplayWidth && py >= 0 && py < graphic.DisplayWidth {
				offset := (py*graphic.DisplayWidth + px) * 3
				r.currBuffer[offset] = color[0]
				r.currBuffer[offset+1] = color[1]
				r.currBuffer[offset+2] = color[2]
			}
		}
	}
}

// ComputeDiff finds changed pixels grouped by color
// Returns a map of color to list of points that changed to that color
func (r *Renderer) ComputeDiff() map[graphic.Color][]graphic.Point {
	diff := make(map[graphic.Color][]graphic.Point)

	for y := 0; y < graphic.DisplayWidth; y++ {
		for x := 0; x < graphic.DisplayWidth; x++ {
			offset := (y*graphic.DisplayWidth + x) * 3
			prevR, prevG, prevB := r.prevBuffer[offset], r.prevBuffer[offset+1], r.prevBuffer[offset+2]
			currR, currG, currB := r.currBuffer[offset], r.currBuffer[offset+1], r.currBuffer[offset+2]

			if prevR != currR || prevG != currG || prevB != currB {
				color := graphic.Color{currR, currG, currB}
				diff[color] = append(diff[color], graphic.Point{X: x, Y: y})
			}
		}
	}

	return diff
}

// Flush sends changed pixels to the device using multi-pixel packets
func (r *Renderer) Flush() error {
	diff := r.ComputeDiff()

	for color, points := range diff {
		// Split points into chunks of MaxPixelsPerPacket
		for i := 0; i < len(points); i += protocol.MaxPixelsPerPacket {
			end := i + protocol.MaxPixelsPerPacket
			if end > len(points) {
				end = len(points)
			}
			chunk := points[i:end]

			if err := protocol.SetPixels(r.device, color, chunk); err != nil {
				return err
			}
			time.Sleep(protocol.PacketDelay)
		}
	}

	// Update previous buffer
	copy(r.prevBuffer[:], r.currBuffer[:])

	return nil
}

// SetPrevBuffer sets the previous buffer (used for initial state)
func (r *Renderer) SetPrevBuffer(data []byte) {
	copy(r.prevBuffer[:], data)
}

// SetCurrBuffer sets the current buffer (used for initial state)
func (r *Renderer) SetCurrBuffer(data []byte) {
	copy(r.currBuffer[:], data)
}

// GetCurrBuffer returns a copy of the current buffer (for testing)
func (r *Renderer) GetCurrBuffer() []byte {
	buf := make([]byte, len(r.currBuffer))
	copy(buf, r.currBuffer[:])
	return buf
}

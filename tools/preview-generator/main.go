// Preview generator for README GIFs.
// See AGENT.md for regeneration guidelines.

package main

import (
	"fmt"
	"image"
	"image/gif"
	"io"
	"os"
	"path/filepath"

	"github.com/pracucci/idotmatrix-overclocked/pkg/fire"
	"github.com/pracucci/idotmatrix-overclocked/pkg/games/snake"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/grot"
	"github.com/pracucci/idotmatrix-overclocked/pkg/text"
)


const previewDir = "pkg/assets/preview"

func main() {
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create %s directory: %v\n", previewDir, err)
		os.Exit(1)
	}

	fmt.Println("Generating text preview...")
	if err := generateTextPreview(filepath.Join(previewDir, "text-preview.gif")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate text preview: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating emoji preview...")
	if err := generateEmojiPreview(filepath.Join(previewDir, "emoji-preview.gif")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate emoji preview: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating grot preview...")
	if err := generateGrotPreview(filepath.Join(previewDir, "grot-preview.gif")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate grot preview: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating snake preview...")
	if err := generateSnakePreview(filepath.Join(previewDir, "snake-preview.gif")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate snake preview: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating fire preview...")
	if err := generateFirePreview(filepath.Join(previewDir, "fire-preview.gif")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate fire preview: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All previews generated successfully!")
}

// generateTextPreview creates a text animation preview using fireworks with "FIRE!" in red.
func generateTextPreview(outputPath string) error {
	opts := text.DefaultAnimationOptions()
	opts.TextColor = graphic.Red
	opts.ShadowColor = graphic.DarkRed
	img := text.GenerateFireworksText("FIRE!", opts)

	gifBytes, err := img.GIFBytes()
	if err != nil {
		return fmt.Errorf("failed to encode text GIF: %w", err)
	}

	return os.WriteFile(outputPath, gifBytes, 0644)
}

// generateEmojiPreview copies the party emoji GIF to the output path.
func generateEmojiPreview(outputPath string) error {
	return copyFile(filepath.Join("pkg", "assets", "emoji", "party.gif"), outputPath)
}

// generateGrotPreview generates the matrix animation GIF.
func generateGrotPreview(outputPath string) error {
	img, err := grot.GenerateMatrix()
	if err != nil {
		return fmt.Errorf("failed to generate matrix: %w", err)
	}

	gifBytes, err := img.GIFBytes()
	if err != nil {
		return fmt.Errorf("failed to encode grot GIF: %w", err)
	}

	return os.WriteFile(outputPath, gifBytes, 0644)
}

// generateFirePreview generates a DOOM-style fire animation.
func generateFirePreview(outputPath string) error {
	gifData := fire.GenerateGIF()
	return os.WriteFile(outputPath, gifData, 0644)
}

// generateSnakePreview creates a snake game preview GIF with three phases:
// 1. Cover image (2s)
// 2. Level interstitial (appearing text)
// 3. Gameplay simulation (2s)
func generateSnakePreview(outputPath string) error {
	var frames []*image.Paletted
	var delays []int

	// Phase 1: Cover image (1 second = 100 centiseconds)
	coverBuf := snake.GenerateCoverImage()
	coverFrame := graphic.RGBToPaletted(coverBuf)
	frames = append(frames, coverFrame)
	delays = append(delays, 100) // 1 second

	// Phase 2: Level interstitial - "LEVEL 1" appearing letter-by-letter
	levelOpts := text.DefaultAnimationOptions()
	levelOpts.TextColor = graphic.Green
	levelOpts.ShadowColor = graphic.DarkGreen
	levelOpts.Background = graphic.Black
	levelOpts.LetterDelay = 20 // 200ms per letter
	levelOpts.HoldDelay = 100  // 1s hold

	levelFrames := text.GenerateAppearingFrames("LEVEL 1", levelOpts)
	for _, frame := range levelFrames {
		frames = append(frames, graphic.RGBToPaletted(frame.Data))
		delays = append(delays, frame.Delay)
	}

	// Phase 3: Gameplay simulation (2 seconds at 100ms per frame = 20 frames)
	gameMap := snake.NewMap() // Empty map for level 1

	// Simulate snake moving right
	// Snake starts at center-ish and moves right
	snakeBody := []struct{ x, y int }{
		{32, 32}, // Head
		{31, 32},
		{30, 32},
		{29, 32},
		{28, 32}, // Tail
	}
	foodX, foodY := 45, 32 // Food position

	for i := 0; i < 20; i++ {
		// Generate fresh background for each frame and brighten for preview visibility
		frameBuf := snake.GenerateBackgroundWithObstacles(gameMap)
		brightenBackground(frameBuf)

		// Draw food (red)
		setPixelRGB(frameBuf, foodX, foodY, snake.Red[0], snake.Red[1], snake.Red[2])

		// Draw snake
		for j, segment := range snakeBody {
			var col [3]uint8
			if j == 0 {
				col = snake.BrightGreen // Head
			} else {
				col = snake.Green // Body
			}
			setPixelRGB(frameBuf, segment.x, segment.y, col[0], col[1], col[2])
		}

		frames = append(frames, graphic.RGBToPaletted(frameBuf))
		delays = append(delays, 10) // 100ms per frame

		// Move snake right (shift all segments)
		for j := len(snakeBody) - 1; j > 0; j-- {
			snakeBody[j].x = snakeBody[j-1].x
			snakeBody[j].y = snakeBody[j-1].y
		}
		snakeBody[0].x++ // Move head right
	}

	// Encode and write GIF
	g := &gif.GIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: 0, // Loop forever
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := gif.EncodeAll(f, g); err != nil {
		return fmt.Errorf("failed to encode GIF: %w", err)
	}

	return nil
}

// setPixelRGB sets a single pixel in the RGB image buffer.
func setPixelRGB(buf []byte, x, y int, r, g, b uint8) {
	if x < 0 || x >= graphic.DisplayWidth || y < 0 || y >= graphic.DisplayHeight {
		return
	}
	offset := (y*graphic.DisplayWidth + x) * 3
	buf[offset] = r
	buf[offset+1] = g
	buf[offset+2] = b
}

// brightenBackground increases the brightness of dark terrain colors for better preview visibility.
// The actual game uses very dark browns that may not display well in GIFs.
func brightenBackground(buf []byte) {
	for i := 0; i < len(buf); i += 3 {
		r, g, b := buf[i], buf[i+1], buf[i+2]
		// Brighten dark browns (terrain colors are around 45-50, 30-35, 18-20)
		if r < 60 && g < 50 && b < 30 {
			buf[i] = min(255, r*3)
			buf[i+1] = min(255, g*3)
			buf[i+2] = min(255, b*3)
		}
	}
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

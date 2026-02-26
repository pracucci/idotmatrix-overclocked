// emoji-generator resizes animated GIF emoji images to 64x64.
// The generated GIFs are stored in pkg/emoji/assets/ and embedded into the binary.
//
// Usage: go run ./cmd/emoji-generator
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"path/filepath"
)

const (
	srcSize = 512
	dstSize = 64
)

func main() {
	assetsDir := "pkg/emoji/assets"

	emojis := []string{
		"thumbsup", "thumbsdown", "hearthands", "clap", "joy",
		"rofl", "party", "scream", "rage", "scared", "mindblow",
		"coldface", "hotface", "robot", "sparkles", "tada",
		"100", "confetti", "risinghands", "rocket", "birthday",
	}

	for _, name := range emojis {
		srcPath := filepath.Join(assetsDir, name+"_512.gif")
		dstPath := filepath.Join(assetsDir, name+".gif")

		// Skip if source doesn't exist (already processed)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("Skipped %s (no source file)\n", name)
			continue
		}

		if err := resizeGIF(srcPath, dstPath); err != nil {
			fmt.Printf("Failed to process %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s\n", dstPath)
	}

	// Clean up source files
	for _, name := range emojis {
		os.Remove(filepath.Join(assetsDir, name+"_512.gif"))
	}

	// Remove old PNG files
	oldFiles := []string{"thumbsup.png", "thumbsdown.png", "pray.png", "lol.png"}
	for _, f := range oldFiles {
		os.Remove(filepath.Join(assetsDir, f))
	}

	fmt.Println("Done!")
}

func resizeGIF(srcPath, dstPath string) error {
	// Open source GIF
	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open GIF: %w", err)
	}
	defer f.Close()

	srcGIF, err := gif.DecodeAll(f)
	if err != nil {
		return fmt.Errorf("failed to decode GIF: %w", err)
	}

	// Create destination GIF
	dstGIF := &gif.GIF{
		Image:     make([]*image.Paletted, len(srcGIF.Image)),
		Delay:     srcGIF.Delay,
		LoopCount: srcGIF.LoopCount,
	}

	// Process each frame
	// We need to handle GIF's cumulative frame rendering
	canvas := image.NewRGBA(image.Rect(0, 0, srcSize, srcSize))

	// Fill canvas with black initially
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.Point{}, draw.Src)

	for i, srcFrame := range srcGIF.Image {
		// Draw frame onto canvas (handles disposal and frame positioning)
		bounds := srcFrame.Bounds()
		draw.Draw(canvas, bounds, srcFrame, bounds.Min, draw.Over)

		// Resize canvas to 64x64 using bilinear interpolation
		resized := resizeRGBA(canvas, dstSize, dstSize)

		// Convert to paletted image
		palette := buildPalette(resized)
		paletted := image.NewPaletted(image.Rect(0, 0, dstSize, dstSize), palette)
		draw.FloydSteinberg.Draw(paletted, paletted.Bounds(), resized, image.Point{})

		dstGIF.Image[i] = paletted

		// Handle disposal
		if i < len(srcGIF.Disposal) {
			switch srcGIF.Disposal[i] {
			case gif.DisposalBackground:
				// Clear the canvas
				draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.Point{}, draw.Src)
			}
		}
	}

	// Write destination GIF
	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create GIF file: %w", err)
	}
	defer out.Close()

	if err := gif.EncodeAll(out, dstGIF); err != nil {
		return fmt.Errorf("failed to encode GIF: %w", err)
	}

	return nil
}

// resizeRGBA resizes an RGBA image using bilinear interpolation
func resizeRGBA(src *image.RGBA, newWidth, newHeight int) *image.RGBA {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	xRatio := float64(srcW) / float64(newWidth)
	yRatio := float64(srcH) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Map destination pixel to source coordinates
			srcX := float64(x) * xRatio
			srcY := float64(y) * yRatio

			// Bilinear interpolation
			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			if x1 >= srcW {
				x1 = srcW - 1
			}
			if y1 >= srcH {
				y1 = srcH - 1
			}

			xFrac := srcX - float64(x0)
			yFrac := srcY - float64(y0)

			// Get the four surrounding pixels
			c00 := src.RGBAAt(x0+srcBounds.Min.X, y0+srcBounds.Min.Y)
			c10 := src.RGBAAt(x1+srcBounds.Min.X, y0+srcBounds.Min.Y)
			c01 := src.RGBAAt(x0+srcBounds.Min.X, y1+srcBounds.Min.Y)
			c11 := src.RGBAAt(x1+srcBounds.Min.X, y1+srcBounds.Min.Y)

			// Interpolate
			r := bilinear(float64(c00.R), float64(c10.R), float64(c01.R), float64(c11.R), xFrac, yFrac)
			g := bilinear(float64(c00.G), float64(c10.G), float64(c01.G), float64(c11.G), xFrac, yFrac)
			b := bilinear(float64(c00.B), float64(c10.B), float64(c01.B), float64(c11.B), xFrac, yFrac)
			a := bilinear(float64(c00.A), float64(c10.A), float64(c01.A), float64(c11.A), xFrac, yFrac)

			dst.SetRGBA(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}

	return dst
}

func bilinear(c00, c10, c01, c11, xFrac, yFrac float64) float64 {
	top := c00*(1-xFrac) + c10*xFrac
	bottom := c01*(1-xFrac) + c11*xFrac
	return top*(1-yFrac) + bottom*yFrac
}

// buildPalette extracts up to 256 colors from the image
func buildPalette(img *image.RGBA) color.Palette {
	colorSet := make(map[color.RGBA]struct{})
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			colorSet[c] = struct{}{}
		}
	}

	palette := make(color.Palette, 0, 256)

	// Always include black first (for background)
	palette = append(palette, color.RGBA{0, 0, 0, 255})
	delete(colorSet, color.RGBA{0, 0, 0, 255})

	for c := range colorSet {
		if len(palette) >= 256 {
			break
		}
		palette = append(palette, c)
	}

	return palette
}

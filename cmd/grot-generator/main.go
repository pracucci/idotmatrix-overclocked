// grot-generator downloads and resizes animated GIF images to 64x64.
// The generated GIFs are stored in pkg/grot/assets/ and embedded into the binary.
//
// Usage: go run ./cmd/grot-generator
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	dstSize = 64
)

type grotSource struct {
	Name string
	URL  string
}

// Use GIF URLs instead of webp (Giphy serves both formats)
var sources = []grotSource{
	{Name: "halloween-1", URL: "https://media0.giphy.com/media/biSOsz3RKmYaOzorSZ/giphy.gif"},
	{Name: "halloween-2", URL: "https://media1.giphy.com/media/nq4DrCLIrgvp5Tgc3c/giphy.gif"},
	{Name: "halloween-3", URL: "https://media0.giphy.com/media/MKmln6IFypmLQxSSxx/giphy.gif"},
	{Name: "halloween-4", URL: "https://media0.giphy.com/media/AeWIgu9YwGZLcZYJC0/giphy.gif"},
	{Name: "halloween-5", URL: "https://media1.giphy.com/media/fRxhvwWa7cyDJ66FKS/giphy.gif"},
	{Name: "halloween-6", URL: "https://media4.giphy.com/media/D5fr17cz4qfoRFqYCu/giphy.gif"},
	{Name: "halloween-7", URL: "https://media1.giphy.com/media/YjWmYzUfe5TEqb8cNE/giphy.gif"},
}

func main() {
	assetsDir := "pkg/grot/assets"

	// Create assets directory if it doesn't exist
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		fmt.Printf("Failed to create assets directory: %v\n", err)
		os.Exit(1)
	}

	for _, src := range sources {
		tempGifPath := filepath.Join(assetsDir, src.Name+"_temp.gif")
		dstPath := filepath.Join(assetsDir, src.Name+".gif")

		// Download GIF
		fmt.Printf("Downloading %s...\n", src.Name)
		if err := downloadFile(src.URL, tempGifPath); err != nil {
			fmt.Printf("Failed to download %s: %v\n", src.Name, err)
			os.Exit(1)
		}

		// Resize GIF to 64x64
		fmt.Printf("Resizing %s to 64x64...\n", src.Name)
		if err := resizeGIF(tempGifPath, dstPath); err != nil {
			fmt.Printf("Failed to resize %s: %v\n", src.Name, err)
			os.Exit(1)
		}

		// Clean up temp files
		os.Remove(tempGifPath)

		fmt.Printf("Generated %s\n", dstPath)
	}

	fmt.Println("Done!")
}

func downloadFile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
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

	// Get source dimensions
	srcW := srcGIF.Config.Width
	srcH := srcGIF.Config.Height
	if srcW == 0 || srcH == 0 {
		if len(srcGIF.Image) > 0 {
			srcW = srcGIF.Image[0].Bounds().Dx()
			srcH = srcGIF.Image[0].Bounds().Dy()
		}
	}

	// Create destination GIF
	dstGIF := &gif.GIF{
		Image:     make([]*image.Paletted, len(srcGIF.Image)),
		Delay:     srcGIF.Delay,
		LoopCount: srcGIF.LoopCount,
	}

	// Process each frame
	canvas := image.NewRGBA(image.Rect(0, 0, srcW, srcH))

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

package snake

// Color palette
var (
	black       = [3]uint8{0, 0, 0}
	darkGreen   = [3]uint8{0, 100, 0}
	green       = [3]uint8{0, 200, 0}
	brightGreen = [3]uint8{50, 255, 50}
	darkRed     = [3]uint8{100, 0, 0}
	red         = [3]uint8{255, 0, 0}
	white       = [3]uint8{255, 255, 255}
	gray        = [3]uint8{80, 80, 80}

	// Obstacle colors
	rockColor     = [3]uint8{55, 55, 60}    // Dark gray with slight blue tint
	rockColorAlt  = [3]uint8{50, 50, 55}    // Slightly darker variation
	lakeColor     = [3]uint8{35, 70, 135}   // Blue
	lakeColorAlt  = [3]uint8{40, 80, 145}   // Lighter wave variation
)

// Brown palette - limited to 2 colors due to display constraints
// (channels < 28 render as black)
var brownPalette = [][3]uint8{
	{45, 31, 18}, // darker brown (column 20)
	{49, 34, 19}, // lighter brown (column 24)
}

// 5x7 pixel font data for uppercase letters and some symbols
var font5x7 = map[rune][7]uint8{
	'S': {0x1E, 0x01, 0x01, 0x0E, 0x10, 0x10, 0x0F},
	'N': {0x11, 0x13, 0x15, 0x19, 0x11, 0x11, 0x11},
	'A': {0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'K': {0x11, 0x09, 0x05, 0x03, 0x05, 0x09, 0x11},
	'E': {0x1F, 0x01, 0x01, 0x0F, 0x01, 0x01, 0x1F},
	'G': {0x0E, 0x11, 0x01, 0x1D, 0x11, 0x11, 0x0E},
	'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
	'O': {0x0E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
	'V': {0x11, 0x11, 0x11, 0x11, 0x11, 0x0A, 0x04},
	'R': {0x0F, 0x11, 0x11, 0x0F, 0x05, 0x09, 0x11},
	'=': {0x00, 0x00, 0x1F, 0x00, 0x1F, 0x00, 0x00},
	'T': {0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04},
	'Y': {0x11, 0x11, 0x0A, 0x04, 0x04, 0x04, 0x04},
}

// setPixel sets a single pixel in the RGB image buffer
func setPixel(img []byte, x, y int, color [3]uint8) {
	if x < 0 || x >= 64 || y < 0 || y >= 64 {
		return
	}
	offset := (y*64 + x) * 3
	img[offset] = color[0]
	img[offset+1] = color[1]
	img[offset+2] = color[2]
}

// drawRect draws a filled rectangle
func drawRect(img []byte, x, y, w, h int, color [3]uint8) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			setPixel(img, x+dx, y+dy, color)
		}
	}
}

// drawChar draws a single character at the given position using 5x7 font
func drawChar(img []byte, char rune, x, y int, color [3]uint8) {
	data, ok := font5x7[char]
	if !ok {
		return
	}
	for row := 0; row < 7; row++ {
		for col := 0; col < 5; col++ {
			if data[row]&(1<<col) != 0 {
				setPixel(img, x+col, y+row, color)
			}
		}
	}
}

// drawText draws a string of text at the given position
func drawText(img []byte, text string, x, y int, color [3]uint8) {
	for i, char := range text {
		drawChar(img, char, x+i*6, y, color)
	}
}

// drawCoiledSnake draws a coiled snake sprite at the given center position
func drawCoiledSnake(img []byte, cx, cy int) {
	// Draw a spiral/coiled snake body
	// Outer ring
	coilPoints := []struct{ x, y int }{
		// Outer coil (clockwise)
		{-8, -4}, {-7, -5}, {-6, -6}, {-5, -7}, {-4, -7}, {-3, -7}, {-2, -7}, {-1, -7},
		{0, -7}, {1, -7}, {2, -7}, {3, -7}, {4, -7}, {5, -6}, {6, -5}, {7, -4},
		{7, -3}, {7, -2}, {7, -1}, {7, 0}, {7, 1}, {7, 2}, {7, 3}, {6, 4},
		{5, 5}, {4, 6}, {3, 6}, {2, 6}, {1, 6}, {0, 6}, {-1, 6}, {-2, 6},
		{-3, 6}, {-4, 5}, {-5, 4}, {-5, 3}, {-5, 2}, {-5, 1}, {-5, 0},
		// Inner coil
		{-5, -1}, {-4, -2}, {-3, -3}, {-2, -4}, {-1, -4}, {0, -4}, {1, -4}, {2, -4},
		{3, -3}, {4, -2}, {4, -1}, {4, 0}, {4, 1}, {3, 2}, {2, 3}, {1, 3},
		{0, 3}, {-1, 3}, {-2, 2}, {-2, 1}, {-2, 0},
		// Center coil
		{-1, -1}, {0, -1}, {1, -1}, {1, 0}, {0, 0},
	}

	// Draw body segments with alternating shades
	for i, p := range coilPoints {
		var color [3]uint8
		if i%3 == 0 {
			color = darkGreen
		} else {
			color = green
		}
		setPixel(img, cx+p.x, cy+p.y, color)
		// Make snake thicker
		setPixel(img, cx+p.x+1, cy+p.y, color)
	}

	// Draw head at the end of the coil (outer position)
	headX, headY := cx-8, cy-3
	setPixel(img, headX, headY, brightGreen)
	setPixel(img, headX, headY-1, brightGreen)
	setPixel(img, headX-1, headY, brightGreen)
	setPixel(img, headX-1, headY-1, brightGreen)

	// Eyes
	setPixel(img, headX-1, headY-1, white)
	setPixel(img, headX, headY-1, white)

	// Tongue
	setPixel(img, headX-2, headY, red)
	setPixel(img, headX-3, headY-1, red)
	setPixel(img, headX-3, headY+1, red)
}

// drawDeadSnake draws a flat dead snake with X eyes
func drawDeadSnake(img []byte, cx, cy int) {
	// Draw a wavy flat snake body
	bodyPoints := []struct{ x, y int }{
		{-20, 2}, {-19, 1}, {-18, 0}, {-17, 0}, {-16, 1}, {-15, 2}, {-14, 2},
		{-13, 1}, {-12, 0}, {-11, 0}, {-10, 1}, {-9, 2}, {-8, 2},
		{-7, 1}, {-6, 0}, {-5, 0}, {-4, 0}, {-3, 0}, {-2, 0}, {-1, 0}, {0, 0},
		{1, 0}, {2, 0}, {3, 0}, {4, 0}, {5, 0}, {6, 0},
	}

	// Draw body with alternating shades
	for i, p := range bodyPoints {
		var color [3]uint8
		if i%2 == 0 {
			color = darkGreen
		} else {
			color = green
		}
		setPixel(img, cx+p.x, cy+p.y, color)
		setPixel(img, cx+p.x, cy+p.y+1, color)
	}

	// Draw head
	headX, headY := cx+8, cy
	drawRect(img, headX, headY-1, 4, 4, green)

	// X eyes (dead)
	setPixel(img, headX, headY-1, red)
	setPixel(img, headX+1, headY, red)
	setPixel(img, headX+1, headY-1, red)
	setPixel(img, headX, headY, red)

	setPixel(img, headX+2, headY-1, red)
	setPixel(img, headX+3, headY, red)
	setPixel(img, headX+3, headY-1, red)
	setPixel(img, headX+2, headY, red)
}

// GenerateCoverImage creates the title screen with "SNAKE" text and coiled snake
func GenerateCoverImage() []byte {
	img := make([]byte, 64*64*3)

	// Dark gradient background with subtle pattern
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			// Subtle dark gradient from top to bottom
			intensity := uint8(10 + y/8)
			// Add some subtle noise/pattern
			if (x+y)%8 == 0 {
				intensity += 5
			}
			setPixel(img, x, y, [3]uint8{0, intensity / 2, 0})
		}
	}

	// Draw decorative corner dots
	corners := [][2]int{{2, 2}, {61, 2}, {2, 61}, {61, 61}}
	for _, c := range corners {
		setPixel(img, c[0], c[1], darkGreen)
		setPixel(img, c[0]-1, c[1], darkGreen)
		setPixel(img, c[0]+1, c[1], darkGreen)
		setPixel(img, c[0], c[1]-1, darkGreen)
		setPixel(img, c[0], c[1]+1, darkGreen)
	}

	// Draw "SNAKE" title at top (centered)
	// "SNAKE" is 5 chars * 6 pixels = 30 pixels wide, center at (64-30)/2 = 17
	// Draw shadow first
	drawText(img, "SNAKE", 18, 7, [3]uint8{0, 50, 0})
	// Draw main text
	drawText(img, "SNAKE", 17, 6, brightGreen)

	// Draw coiled snake in center
	drawCoiledSnake(img, 32, 38)

	// Draw small decorative border lines
	for x := 5; x < 59; x++ {
		if x%2 == 0 {
			setPixel(img, x, 58, gray)
		}
	}

	return img
}

// seed represents a point for Voronoi region generation
type seed struct {
	x, y       int
	colorIndex int
}

// generateVoronoiBackground creates a terrain-like background for gameplay
// using Voronoi-like regions with 2 alternating brown colors
func generateVoronoiBackground() []byte {
	img := make([]byte, 64*64*3)

	// Generate random seed points for Voronoi regions
	numSeeds := 40
	seeds := make([]seed, numSeeds)

	// Use a simple LCG for deterministic pseudo-random generation
	lcgState := uint32(12345)
	lcgNext := func() uint32 {
		lcgState = lcgState*1103515245 + 12345
		return (lcgState >> 16) & 0x7FFF
	}

	// Generate seed points distributed across the image
	for i := 0; i < numSeeds; i++ {
		seeds[i] = seed{
			x: int(lcgNext() % 64),
			y: int(lcgNext() % 64),
		}
	}

	// First pass: assign each pixel to nearest seed to determine regions
	regionMap := make([]int, 64*64)
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			minDist := 64*64 + 64*64 // max possible squared distance
			nearestSeed := 0
			for i, s := range seeds {
				dx := x - s.x
				dy := y - s.y
				dist := dx*dx + dy*dy
				if dist < minDist {
					minDist = dist
					nearestSeed = i
				}
			}
			regionMap[y*64+x] = nearestSeed
		}
	}

	// Build adjacency list for seeds
	adjacency := make([]map[int]bool, numSeeds)
	for i := range adjacency {
		adjacency[i] = make(map[int]bool)
	}

	// Find adjacent regions by checking neighboring pixels
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			currentRegion := regionMap[y*64+x]
			if x < 63 {
				neighborRegion := regionMap[y*64+x+1]
				if neighborRegion != currentRegion {
					adjacency[currentRegion][neighborRegion] = true
					adjacency[neighborRegion][currentRegion] = true
				}
			}
			if y < 63 {
				neighborRegion := regionMap[(y+1)*64+x]
				if neighborRegion != currentRegion {
					adjacency[currentRegion][neighborRegion] = true
					adjacency[neighborRegion][currentRegion] = true
				}
			}
		}
	}

	// Color regions with 2 colors, alternating so neighbors differ
	// This is a graph 2-coloring problem
	for i := range seeds {
		seeds[i].colorIndex = -1
	}

	for i := 0; i < numSeeds; i++ {
		if seeds[i].colorIndex >= 0 {
			continue
		}
		// BFS to color this connected component
		queue := []int{i}
		seeds[i].colorIndex = 0
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			nextColor := 1 - seeds[current].colorIndex
			for neighbor := range adjacency[current] {
				if seeds[neighbor].colorIndex < 0 {
					seeds[neighbor].colorIndex = nextColor
					queue = append(queue, neighbor)
				}
			}
		}
	}

	// Draw the image
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			region := regionMap[y*64+x]
			colorIdx := seeds[region].colorIndex
			setPixel(img, x, y, brownPalette[colorIdx])
		}
	}

	return img
}

// GenerateBackgroundWithObstacles creates a terrain background with obstacles overlaid.
func GenerateBackgroundWithObstacles(gameMap *Map) []byte {
	// Start with the Voronoi terrain background
	img := generateVoronoiBackground()

	// Overlay obstacles
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			tile := gameMap.GetTile(x, y)
			switch tile {
			case TileRock:
				// Dark gray rock with subtle variation
				color := rockColor
				if (x+y)%2 == 0 {
					color = rockColorAlt
				}
				setPixel(img, x, y, color)
			case TileLake:
				// Blue lake with wave pattern
				color := lakeColor
				if (x+y)%3 == 0 {
					color = lakeColorAlt
				}
				setPixel(img, x, y, color)
			}
		}
	}

	return img
}

// GenerateGameOverImage creates the game over screen
func GenerateGameOverImage() []byte {
	img := make([]byte, 64*64*3)

	// Dark red tinted background
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			intensity := uint8(20 + y/4)
			setPixel(img, x, y, [3]uint8{intensity / 2, 0, 0})
		}
	}

	// Draw "GAME" and "OVER" text centered on screen
	// Each word is 4 chars * 6 pixels = 24 pixels wide, center at (64-24)/2 = 20
	// Total height: 7 + 3 gap + 7 = 17 pixels, center at (64-17)/2 = 23
	shadowColor := [3]uint8{50, 0, 0}
	drawText(img, "GAME", 21, 25, shadowColor)
	drawText(img, "GAME", 20, 24, red)
	drawText(img, "OVER", 21, 35, shadowColor)
	drawText(img, "OVER", 20, 34, red)

	return img
}

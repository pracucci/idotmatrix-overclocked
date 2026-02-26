package snake

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
)

// Direction represents the snake's movement direction.
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// GameState represents the current state of the game.
type GameState string

const (
	StateCover    GameState = "cover"
	StatePlaying  GameState = "playing"
	StateGameOver GameState = "gameover"
	StateEnded    GameState = "ended"
)

// PixelChange represents a pixel update to be sent to the device.
type PixelChange struct {
	pos     Point
	r, g, b uint8
}

// Game represents the snake game state.
type Game struct {
	device      protocol.DeviceConnection
	snake       []Point // Head is snake[0], tail is snake[len-1]
	direction   Direction
	food        Point
	score       int
	running     bool
	gameOver    bool
	state       GameState
	input       InputSource
	background  []byte // Store background RGB data for pixel restoration

	// Level system
	startLevel   int         // Starting level (for testing)
	currentLevel int         // Current level (1-based)
	applesEaten  int         // Apples eaten in current level
	gameMap      *Map        // Current map with obstacles
	levelConfig  LevelConfig // Current level configuration
	growthQueue  int         // Pending growth segments
}

// NewGame creates a new snake game instance.
func NewGame(device protocol.DeviceConnection, startLevel int) *Game {
	if startLevel < 1 {
		startLevel = 1
	}
	g := &Game{
		device:       device,
		running:      true,
		state:        StateCover,
		startLevel:   startLevel,
		currentLevel: startLevel,
	}
	return g
}

// GetState returns the current game state.
func (g *Game) GetState() GameState {
	return g.state
}

// GetScore returns the current score.
func (g *Game) GetScore() int {
	return g.score
}

// GetLevel returns the current level.
func (g *Game) GetLevel() int {
	return g.currentLevel
}

// IsRunning returns whether the game is still running.
func (g *Game) IsRunning() bool {
	return g.running
}

// Stop stops the game.
func (g *Game) Stop() {
	g.running = false
	g.state = StateEnded
}

// SendInput sends a key input to the game.
func (g *Game) SendInput(key string) {
	if g.input == nil {
		return
	}

	channelInput, ok := g.input.(*ChannelInput)
	if !ok {
		return
	}

	switch key {
	case "up":
		channelInput.SendKey('w')
	case "down":
		channelInput.SendKey('s')
	case "left":
		channelInput.SendKey('a')
	case "right":
		channelInput.SendKey('d')
	case "quit":
		channelInput.SendKey('q')
	case "restart":
		channelInput.SendKey('r')
	}
}

// reset initializes the game state for a new game.
func (g *Game) reset() {
	g.currentLevel = g.startLevel
	g.applesEaten = 0
	g.score = 0
	g.gameOver = false
	g.growthQueue = 0
	g.resetSnakePosition(InitialLength)
}

// resetSnakePosition resets the snake to the center with the given length.
func (g *Game) resetSnakePosition(length int) {
	g.snake = make([]Point, length)
	startX := DisplaySize / 2
	startY := DisplaySize / 2
	for i := 0; i < length; i++ {
		g.snake[i] = Point{X: startX - i, Y: startY}
	}
	g.direction = Right
}

// advanceLevel advances to the next level, preserving snake length.
func (g *Game) advanceLevel() {
	g.currentLevel++
	g.applesEaten = 0
	// Preserve snake length when advancing levels
	currentLength := len(g.snake)
	g.resetSnakePosition(currentLength)
}

// setupLevel generates the map and prepares for the current level.
func (g *Game) setupLevel() {
	g.levelConfig = GetLevelConfig(g.currentLevel)

	// Generate new map with obstacles
	mapGen := NewMapGenerator(time.Now().UnixNano())
	g.gameMap = mapGen.Generate(g.levelConfig.NumRocks, g.levelConfig.NumLakes)

	// Generate and store background with obstacles
	g.background = GenerateBackgroundWithObstacles(g.gameMap)
}

// spawnFood places food on a valid terrain position.
func (g *Game) spawnFood() {
	// Get all terrain positions
	terrain := g.gameMap.TerrainPositions()

	// Filter out snake positions
	snakeSet := make(map[Point]bool)
	for _, p := range g.snake {
		snakeSet[p] = true
	}

	var validPositions []Point
	for _, p := range terrain {
		// Skip if on snake
		if snakeSet[p] {
			continue
		}
		// Skip edge positions (at least 1 pixel from edge)
		if p.X == 0 || p.X == DisplaySize-1 || p.Y == 0 || p.Y == DisplaySize-1 {
			continue
		}
		validPositions = append(validPositions, p)
	}

	if len(validPositions) == 0 {
		// No valid positions (extremely rare), just pick random terrain
		g.food = terrain[rand.Intn(len(terrain))]
		return
	}

	g.food = validPositions[rand.Intn(len(validPositions))]
}

// getBackgroundPixel returns the background color at the given position.
func (g *Game) getBackgroundPixel(x, y int) (uint8, uint8, uint8) {
	offset := (y*64 + x) * 3
	return g.background[offset], g.background[offset+1], g.background[offset+2]
}

// calculateNewHead returns the new head position based on direction.
func (g *Game) calculateNewHead() Point {
	head := g.snake[0]
	switch g.direction {
	case Up:
		return Point{X: head.X, Y: head.Y - 1}
	case Down:
		return Point{X: head.X, Y: head.Y + 1}
	case Left:
		return Point{X: head.X - 1, Y: head.Y}
	case Right:
		return Point{X: head.X + 1, Y: head.Y}
	}
	return head
}

// isCollision checks if the given point causes a collision.
func (g *Game) isCollision(p Point) bool {
	// Wall collision
	if p.X < 0 || p.X >= DisplaySize || p.Y < 0 || p.Y >= DisplaySize {
		return true
	}
	// Obstacle collision
	if g.gameMap.IsObstacle(p.X, p.Y) {
		return true
	}
	// Self collision (check against body, not including the tail that will be removed)
	for i := 0; i < len(g.snake)-1; i++ {
		if g.snake[i] == p {
			return true
		}
	}
	return false
}

// move moves the snake and returns pixel changes to render.
// Returns true if level should advance.
func (g *Game) move() ([]PixelChange, bool) {
	var changes []PixelChange

	newHead := g.calculateNewHead()

	if g.isCollision(newHead) {
		g.gameOver = true
		return nil, false
	}

	// Add new head (green pixel)
	changes = append(changes, PixelChange{newHead, 0, 255, 0})
	g.snake = append([]Point{newHead}, g.snake...)

	// Check if eating food
	if newHead == g.food {
		g.score++
		g.applesEaten++
		g.growthQueue += GrowthPerApple // Queue growth

		// Check for level advancement
		if g.applesEaten >= ApplesPerLevel {
			return changes, true // Signal level advance
		}

		// Spawn new food
		g.spawnFood()
		// Draw new food (red pixel)
		changes = append(changes, PixelChange{g.food, 255, 0, 0})
	}

	// Handle tail: grow if growth queued, otherwise remove tail
	if g.growthQueue > 0 {
		g.growthQueue--
		// Don't remove tail when growing
	} else {
		// Remove tail (restore background pixel)
		tail := g.snake[len(g.snake)-1]
		r, gb, b := g.getBackgroundPixel(tail.X, tail.Y)
		changes = append(changes, PixelChange{tail, r, gb, b})
		g.snake = g.snake[:len(g.snake)-1]
	}

	return changes, false
}

// handleInput processes keyboard input.
func (g *Game) handleInput() {
	if g.input == nil {
		return
	}

	key, ok := g.input.TryReadKey()
	if !ok {
		return
	}

	switch key {
	case 'w', 'W':
		if g.direction != Down {
			g.direction = Up
		}
	case 's', 'S':
		if g.direction != Up {
			g.direction = Down
		}
	case 'a', 'A':
		if g.direction != Right {
			g.direction = Left
		}
	case 'd', 'D':
		if g.direction != Left {
			g.direction = Right
		}
	case 'q', 'Q':
		g.running = false
	}
}

// renderInitial draws the initial snake and food on the display.
func (g *Game) renderInitial() {
	// Draw the snake
	for _, p := range g.snake {
		protocol.SetPixel(g.device, p.X, p.Y, 0, 255, 0)
		time.Sleep(20 * time.Millisecond)
	}
	// Draw the food
	protocol.SetPixel(g.device, g.food.X, g.food.Y, 255, 0, 0)
	time.Sleep(20 * time.Millisecond)
}

// showImage displays an image on the device.
func (g *Game) showImage(rgbData []byte) error {
	if err := protocol.SetDrawMode(g.device, 1); err != nil {
		return err
	}
	if err := protocol.SendImage(g.device, rgbData); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return nil
}

// waitForKey blocks until a key is pressed.
func (g *Game) waitForKey() rune {
	if g.input == nil {
		return 'q'
	}
	return g.input.ReadKey()
}

// runLevel runs a single level and returns true if the game should continue.
func (g *Game) runLevel() (continueGame bool, advanceLevel bool) {
	// Setup level
	g.setupLevel()

	// Show level interstitial
	if err := g.showLevelInterstitial(); err != nil {
		return false, false
	}

	// Display background with obstacles
	if err := g.showImage(g.background); err != nil {
		return false, false
	}

	// Spawn food and draw initial state
	g.spawnFood()
	g.renderInitial()

	// Game tick loop
	for g.running && !g.gameOver {
		tickStart := time.Now()

		g.handleInput()

		if !g.gameOver {
			changes, shouldAdvance := g.move()
			if shouldAdvance {
				return true, true // Continue game, advance level
			}
			for _, c := range changes {
				protocol.SetPixel(g.device, c.pos.X, c.pos.Y, c.r, c.g, c.b)
				time.Sleep(20 * time.Millisecond)
			}
		}

		elapsed := time.Since(tickStart)
		sleepTime := g.levelConfig.TickDelay - elapsed
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	}

	if !g.running {
		return false, false // Quit
	}

	return true, false // Game over, don't advance
}

// Run starts the main game loop using terminal input.
func (g *Game) Run() error {
	fmt.Println("Starting Snake!")
	fmt.Println("Controls: WASD or Arrow keys to move, Q to quit, R to restart")

	input, err := NewTerminalInput()
	if err != nil {
		fmt.Printf("Warning: could not set raw mode: %v\n", err)
		return err
	}
	defer input.Close()

	g.input = input
	return g.runWithInput()
}

// RunWithChannelInput starts the game loop with a channel-based input source (for HTTP).
func (g *Game) RunWithChannelInput(input *ChannelInput) error {
	g.input = input
	return g.runWithInput()
}

// runWithInput runs the game loop with the current input source.
func (g *Game) runWithInput() error {
	for g.running {
		// Show cover image and wait for key to start
		g.state = StateCover
		if err := g.showImage(GenerateCoverImage()); err != nil {
			return err
		}
		key := g.waitForKey()
		if key == 'q' || key == 'Q' {
			break
		}

		// Reset and start game
		g.reset()
		g.state = StatePlaying

		// Level loop
		for g.running && !g.gameOver {
			continueGame, advance := g.runLevel()
			if !continueGame {
				break
			}
			if advance {
				g.advanceLevel()
			}
		}

		if !g.running {
			break
		}

		// Game over - show game over image and wait for any key to restart (Q to quit)
		g.state = StateGameOver
		if err := g.showImage(GenerateGameOverImage()); err != nil {
			return err
		}
		key = g.waitForKey()
		if key == 'q' || key == 'Q' {
			break
		}
	}

	g.running = false
	g.state = StateEnded
	return nil
}

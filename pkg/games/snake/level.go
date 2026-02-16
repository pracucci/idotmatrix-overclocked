package snake

import "time"

// Game constants
const (
	ApplesPerLevel = 3  // Apples needed to advance to next level
	GrowthPerApple = 3  // Snake grows by 3 pixels per apple
	InitialLength  = 3  // Initial snake length
	DisplaySize    = 64 // Display size in pixels

	// Speed settings (tick delays)
	SlowTickDelay   = 100 * time.Millisecond // Level 1
	MediumTickDelay = 70 * time.Millisecond  // Level 2
	FastTickDelay   = 35 * time.Millisecond  // Level 3+

	// Obstacle caps
	MaxRocks = 20
	MaxLakes = 10
)

// LevelConfig holds the configuration for a specific level.
type LevelConfig struct {
	Level     int
	TickDelay time.Duration
	NumRocks  int
	NumLakes  int
}

// GetLevelConfig returns the configuration for the given level.
// Level is 1-based.
func GetLevelConfig(level int) LevelConfig {
	config := LevelConfig{
		Level: level,
	}

	switch {
	case level == 1:
		// Level 1: Slow, no obstacles
		config.TickDelay = SlowTickDelay
		config.NumRocks = 0
		config.NumLakes = 0

	case level == 2:
		// Level 2: Medium speed, few obstacles
		config.TickDelay = MediumTickDelay
		config.NumRocks = 3
		config.NumLakes = 1

	default:
		// Level 3+: Fast, increasing obstacles
		config.TickDelay = FastTickDelay

		// Increase obstacles with level, but cap them
		config.NumRocks = min(3+(level-2)*2, MaxRocks)
		config.NumLakes = min(1+(level-2), MaxLakes)
	}

	return config
}

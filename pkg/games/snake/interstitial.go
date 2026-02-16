package snake

import (
	"fmt"
	"time"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
	"github.com/pracucci/idotmatrix-overclocked/pkg/protocol"
	"github.com/pracucci/idotmatrix-overclocked/pkg/text"
)

// showLevelInterstitial displays the "LEVEL N" animation.
func (g *Game) showLevelInterstitial() error {
	levelText := fmt.Sprintf("LEVEL %d", g.currentLevel)

	opts := text.DefaultAnimationOptions()
	opts.TextColor = graphic.Green
	opts.ShadowColor = graphic.DarkGreen
	opts.Background = graphic.Black
	opts.LetterDelay = 30  // 300ms per letter (slower for device stability)
	opts.HoldDelay = 150   // 1.5s hold

	frames := text.GenerateAppearingFrames(levelText, opts)
	for _, frame := range frames {
		if err := protocol.SendImage(g.device, frame.Data); err != nil {
			return err
		}
		// Wait for frame delay plus extra buffer for device to process
		time.Sleep(time.Duration(frame.Delay)*10*time.Millisecond + 100*time.Millisecond)
	}

	// Extra delay before transitioning to game
	time.Sleep(500 * time.Millisecond)

	return nil
}

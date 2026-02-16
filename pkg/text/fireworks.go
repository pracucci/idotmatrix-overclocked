package text

import (
	"image"
	"image/gif"
	"math"
	"math/rand"
	"time"

	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// Fireworks physics constants (tuned for 64x64 display)
const (
	fwGravity          = 0.15
	fwLaunchVelMin     = -4.5
	fwLaunchVelMax     = -3.0
	fwExplosionVelMin  = 1.5
	fwExplosionVelMax  = 3.0
	fwParticleCountMin = 20
	fwParticleCountMax = 35
	fwParticleLifeMin  = 20
	fwParticleLifeMax  = 35
	fwMaxFireworks     = 4
	fwSpawnChance      = 0.15
	fwTotalFrames      = 120
	fwFrameDelay       = 3 // 30ms per frame
)

// Firework colors - bright and colorful
var fireworkColors = []graphic.Color{
	graphic.Red,
	graphic.Yellow,
	graphic.Cyan,
	graphic.Magenta,
	graphic.Green,
	graphic.Orange,
	graphic.Blue,
	graphic.White,
	graphic.Pink,
}

// fwParticle represents a single particle in a firework explosion
type fwParticle struct {
	x, y    float64
	vx, vy  float64
	life    int
	maxLife int
	color   graphic.Color
}

// firework represents a single firework (launching or exploding)
type firework struct {
	x, y      float64
	vy        float64
	exploded  bool
	particles []fwParticle
	color     graphic.Color
}

// spawnFirework creates a new firework at the bottom of the screen
func spawnFirework(rng *rand.Rand) *firework {
	return &firework{
		x:        float64(rng.Intn(graphic.DisplayWidth)),
		y:        float64(graphic.DisplayHeight - 1),
		vy:       fwLaunchVelMin + rng.Float64()*(fwLaunchVelMax-fwLaunchVelMin),
		exploded: false,
		color:    fireworkColors[rng.Intn(len(fireworkColors))],
	}
}

// spawnExplodedFirework creates a firework that's already exploded (for seamless loop start)
func spawnExplodedFirework(rng *rand.Rand) *firework {
	fw := &firework{
		x:        float64(10 + rng.Intn(graphic.DisplayWidth-20)),
		y:        float64(10 + rng.Intn(20)),
		exploded: true,
		color:    fireworkColors[rng.Intn(len(fireworkColors))],
	}
	fw.explode(rng)
	// Age the particles randomly so they're mid-explosion
	for i := range fw.particles {
		age := rng.Intn(fw.particles[i].maxLife / 2)
		fw.particles[i].life -= age
		fw.particles[i].x += fw.particles[i].vx * float64(age)
		fw.particles[i].y += fw.particles[i].vy * float64(age)
		// Add gravity effect for aged particles
		fw.particles[i].vy += fwGravity * float64(age)
	}
	return fw
}

// explode converts a launching firework into an explosion of particles
func (fw *firework) explode(rng *rand.Rand) {
	particleCount := fwParticleCountMin + rng.Intn(fwParticleCountMax-fwParticleCountMin+1)
	fw.particles = make([]fwParticle, particleCount)

	for i := 0; i < particleCount; i++ {
		// Radial angle with slight jitter for organic look
		angle := (2 * math.Pi * float64(i) / float64(particleCount)) + (rng.Float64()-0.5)*0.3
		speed := fwExplosionVelMin + rng.Float64()*(fwExplosionVelMax-fwExplosionVelMin)
		life := fwParticleLifeMin + rng.Intn(fwParticleLifeMax-fwParticleLifeMin+1)

		fw.particles[i] = fwParticle{
			x:       fw.x,
			y:       fw.y,
			vx:      math.Cos(angle) * speed,
			vy:      math.Sin(angle) * speed,
			life:    life,
			maxLife: life,
			color:   fw.color,
		}
	}
	fw.exploded = true
}

// update updates the firework state for one frame
func (fw *firework) update(rng *rand.Rand) {
	if !fw.exploded {
		// Launch phase
		fw.y += fw.vy
		fw.vy += fwGravity * 0.3 // Slower gravity during rise

		// Explode when velocity slows or randomly
		if fw.vy >= -0.5 || rng.Float64() < 0.08 {
			fw.explode(rng)
		}
	} else {
		// Explosion phase - update all particles
		aliveParticles := fw.particles[:0]
		for i := range fw.particles {
			p := &fw.particles[i]
			p.x += p.vx
			p.y += p.vy
			p.vy += fwGravity
			p.life--

			if p.life > 0 {
				aliveParticles = append(aliveParticles, *p)
			}
		}
		fw.particles = aliveParticles
	}
}

// isDead returns true if the firework has no more visible elements
func (fw *firework) isDead() bool {
	return fw.exploded && len(fw.particles) == 0
}

// draw renders the firework onto the buffer
func (fw *firework) draw(buf []byte) {
	if !fw.exploded {
		// Draw launch trail
		x, y := int(fw.x), int(fw.y)
		graphic.SetPixel(buf, x, y, fw.color)
		// Fading trail
		trailColor := fadeFireworkColor(fw.color, 1, 2)
		graphic.SetPixel(buf, x, y+1, trailColor)
		dimColor := graphic.Color{trailColor[0] / 2, trailColor[1] / 2, trailColor[2] / 2}
		graphic.SetPixel(buf, x, y+2, dimColor)
	} else {
		// Draw particles
		for _, p := range fw.particles {
			color := fadeFireworkColor(p.color, p.life, p.maxLife)
			graphic.SetPixel(buf, int(p.x), int(p.y), color)
		}
	}
}

// fadeFireworkColor fades a color based on remaining life
func fadeFireworkColor(c graphic.Color, life, maxLife int) graphic.Color {
	if life <= 0 || maxLife <= 0 {
		return graphic.Black
	}
	factor := float64(life) / float64(maxLife)
	return graphic.Color{
		uint8(float64(c[0]) * factor),
		uint8(float64(c[1]) * factor),
		uint8(float64(c[2]) * factor),
	}
}

// GenerateFireworksText creates an animated text display with colorful fireworks.
// The text is displayed centered with fireworks exploding around it.
// LoopCount = 0 (loops forever)
func GenerateFireworksText(text string, opts AnimationOptions) *graphic.Image {
	lines := WrapText(text)

	// Calculate text positioning
	totalHeight := TextBlockHeight(lines)
	startY := (graphic.DisplayHeight - totalHeight) / 2

	// Initialize random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Initialize with some fireworks already in progress for seamless start
	fireworks := []*firework{
		spawnExplodedFirework(rng),
		spawnExplodedFirework(rng),
	}

	var frames []*image.Paletted
	var delays []int

	for frame := 0; frame < fwTotalFrames; frame++ {
		// Create buffer with background
		buf := graphic.NewBufferWithColor(opts.Background)

		// Maybe spawn new firework
		if len(fireworks) < fwMaxFireworks && rng.Float64() < fwSpawnChance {
			fireworks = append(fireworks, spawnFirework(rng))
		}

		// Update all fireworks
		for _, fw := range fireworks {
			fw.update(rng)
		}

		// Remove dead fireworks
		aliveFireworks := fireworks[:0]
		for _, fw := range fireworks {
			if !fw.isDead() {
				aliveFireworks = append(aliveFireworks, fw)
			}
		}
		fireworks = aliveFireworks

		// Draw fireworks BEHIND text
		for _, fw := range fireworks {
			fw.draw(buf)
		}

		// Draw text ON TOP
		if len(lines) == 1 {
			DrawTextCentered(buf, lines[0], opts.TextOptions)
		} else {
			for lineIdx, line := range lines {
				if len(line) == 0 {
					continue
				}
				lineWidth := TextWidth(line)
				x := (graphic.DisplayWidth - lineWidth) / 2
				y := startY + lineIdx*(FontHeight+LineSpacing)
				DrawTextShadowed(buf, line, x, y, opts.TextOptions)
			}
		}

		frames = append(frames, graphic.RGBToPaletted(buf))
		delays = append(delays, fwFrameDelay)
	}

	return &graphic.Image{
		Type: graphic.ImageTypeAnimated,
		GIFData: &gif.GIF{
			Image:     frames,
			Delay:     delays,
			LoopCount: 0, // Loop forever
		},
	}
}

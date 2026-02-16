// Package emoji provides animated emoji images for the iDotMatrix display.
//
// Emoji animations are from Noto Emoji Animation by Google:
// https://googlefonts.github.io/noto-emoji-animation/
//
// Licensed under Apache License 2.0.
package emoji

import (
	"bytes"
	"fmt"
	"image/gif"
	"strings"

	"github.com/pracucci/idotmatrix-overclocked/pkg/assets"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// Emoji defines an emoji with its aliases and asset filename.
type Emoji struct {
	Names    []string // All valid names (lowercase)
	Filename string   // Asset filename (without path)
}

var registry = []Emoji{
	{Names: []string{"thumbsup", "+1"}, Filename: "thumbsup.gif"},
	{Names: []string{"thumbsdown", "-1"}, Filename: "thumbsdown.gif"},
	{Names: []string{"hearthands"}, Filename: "hearthands.gif"},
	{Names: []string{"clap"}, Filename: "clap.gif"},
	{Names: []string{"joy"}, Filename: "joy.gif"},
	{Names: []string{"rofl", "lol"}, Filename: "rofl.gif"},
	{Names: []string{"party"}, Filename: "party.gif"},
	{Names: []string{"scream"}, Filename: "scream.gif"},
	{Names: []string{"rage"}, Filename: "rage.gif"},
	{Names: []string{"scared"}, Filename: "scared.gif"},
	{Names: []string{"mindblow"}, Filename: "mindblow.gif"},
	{Names: []string{"coldface"}, Filename: "coldface.gif"},
	{Names: []string{"hotface"}, Filename: "hotface.gif"},
	{Names: []string{"robot"}, Filename: "robot.gif"},
	{Names: []string{"sparkles"}, Filename: "sparkles.gif"},
	{Names: []string{"tada"}, Filename: "tada.gif"},
	{Names: []string{"100"}, Filename: "100.gif"},
	{Names: []string{"confetti"}, Filename: "confetti.gif"},
	{Names: []string{"risinghands"}, Filename: "risinghands.gif"},
	{Names: []string{"rocket"}, Filename: "rocket.gif"},
	{Names: []string{"birthday"}, Filename: "birthday.gif"},
}

// Lookup finds an emoji by name (case-insensitive).
func Lookup(name string) *Emoji {
	nameLower := strings.ToLower(name)
	for i := range registry {
		for _, n := range registry[i].Names {
			if n == nameLower {
				return &registry[i]
			}
		}
	}
	return nil
}

// Names returns all available emoji names (including aliases).
func Names() []string {
	var names []string
	for _, e := range registry {
		names = append(names, e.Names...)
	}
	return names
}

// Generate creates an animated Image for the given emoji name.
func Generate(name string) (*graphic.Image, error) {
	emoji := Lookup(name)
	if emoji == nil {
		return nil, fmt.Errorf("unknown emoji: %s (available: %s)", name, strings.Join(Names(), ", "))
	}

	data, err := assets.Emoji.ReadFile("emoji/" + emoji.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read emoji asset: %w", err)
	}

	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode emoji GIF: %w", err)
	}

	return &graphic.Image{
		Type:    graphic.ImageTypeAnimated,
		GIFData: g,
	}, nil
}

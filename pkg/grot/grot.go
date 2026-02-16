// Package grot provides animated grot images (Halloween-themed) for the iDotMatrix display.
//
// Grot animations are from Giphy.
package grot

import (
	"bytes"
	"fmt"
	"image/gif"
	"strings"

	"github.com/pracucci/idotmatrix-overclocked/pkg/assets"
	"github.com/pracucci/idotmatrix-overclocked/pkg/graphic"
)

// Grot defines a grot with its aliases and asset filename.
type Grot struct {
	Names    []string // All valid names (lowercase)
	Filename string   // Asset filename (without path)
}

var registry = []Grot{
	{Names: []string{"halloween-1"}, Filename: "halloween-1.gif"},
	{Names: []string{"halloween-2"}, Filename: "halloween-2.gif"},
	{Names: []string{"halloween-3"}, Filename: "halloween-3.gif"},
	{Names: []string{"halloween-4"}, Filename: "halloween-4.gif"},
	{Names: []string{"halloween-5"}, Filename: "halloween-5.gif"},
	{Names: []string{"halloween-6"}, Filename: "halloween-6.gif"},
	{Names: []string{"halloween-7"}, Filename: "halloween-7.gif"},
	{Names: []string{"matrix"}, Filename: ""}, // Procedurally generated
}

// Lookup finds a grot by name (case-insensitive).
func Lookup(name string) *Grot {
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

// Names returns all available grot names (including aliases).
func Names() []string {
	var names []string
	for _, g := range registry {
		names = append(names, g.Names...)
	}
	return names
}

// Generate creates an animated Image for the given grot name.
func Generate(name string) (*graphic.Image, error) {
	g := Lookup(name)
	if g == nil {
		return nil, fmt.Errorf("unknown grot: %s (available: %s)", name, strings.Join(Names(), ", "))
	}

	// Special case for procedurally generated grots
	if g.Filename == "" {
		switch name {
		case "matrix":
			return GenerateMatrix()
		}
		return nil, fmt.Errorf("unknown procedural grot: %s", name)
	}

	data, err := assets.Grot.ReadFile("grot/" + g.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read grot asset: %w", err)
	}

	gifData, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode grot GIF: %w", err)
	}

	return &graphic.Image{
		Type:    graphic.ImageTypeAnimated,
		GIFData: gifData,
	}, nil
}

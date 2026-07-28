package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Entity represents an animated object on screen.
type Entity struct {
	Name string
	Type string

	// Visual representation
	Frames      [][][]rune // [frame][row][col]
	ColorFrames [][][]byte // [frame][row][col] — color mask per frame
	Width       int
	Height      int
	DefStyle    tcell.Style
	TransChar   rune // transparent character (0 = draw all)

	// Position (float for sub-cell precision)
	X, Y float64
	Z    int

	// Velocity per tick
	DX, DY float64

	// Animation
	AnimSpeed float64 // frame advancement per tick
	AnimFrame float64 // current frame index (wrapping)
	AnimTotal float64 // total frame advancement (for DieFrame)

	// Collision
	Physical    bool
	CollDepth   int // entity spans Z to Z+CollDepth
	CollHandler func(*Entity, *Engine)
	Collisions  []*Entity

	// Lifecycle
	DieOffscreen bool
	DieTime      int64 // unix timestamp, 0 = never
	DieFrame     int   // die when AnimTotal >= this, 0 = never
	DeathCB      func(*Entity, *Engine)
	Dead         bool

	// Per-tick callback (replaces default movement)
	Callback func(*Entity, *Engine)
}

func (e *Entity) IntX() int { return int(e.X) }
func (e *Entity) IntY() int { return int(e.Y) }

func (e *Entity) CurrentFrame() [][]rune {
	if len(e.Frames) == 0 {
		return nil
	}
	idx := int(e.AnimFrame) % len(e.Frames)
	if idx < 0 {
		idx = 0
	}
	return e.Frames[idx]
}

func (e *Entity) CurrentColorFrame() [][]byte {
	if len(e.ColorFrames) == 0 {
		return nil
	}
	idx := int(e.AnimFrame) % len(e.ColorFrames)
	if idx < 0 {
		idx = 0
	}
	return e.ColorFrames[idx]
}

func (e *Entity) IsOffscreen(w, h int) bool {
	ix, iy := e.IntX(), e.IntY()
	return ix+e.Width < 0 || ix >= w || iy+e.Height < 0 || iy >= h
}

func (e *Entity) IsSolid(lx, ly int) bool {
	frame := e.CurrentFrame()
	if frame == nil || ly < 0 || ly >= len(frame) || lx < 0 || lx >= len(frame[ly]) {
		return false
	}
	ch := frame[ly][lx]
	if e.TransChar != 0 && ch == e.TransChar {
		return false
	}
	return true
}

// --- Shape parsing ---

// ParseFrames converts multi-line strings into rune grids.
func ParseFrames(shapes []string) ([][][]rune, int, int) {
	var frames [][][]rune
	maxW, maxH := 0, 0
	for _, s := range shapes {
		grid, w, h := parseFrame(s)
		frames = append(frames, grid)
		if w > maxW {
			maxW = w
		}
		if h > maxH {
			maxH = h
		}
	}
	return frames, maxW, maxH
}

func parseFrame(s string) ([][]rune, int, int) {
	lines := strings.Split(s, "\n")
	// Strip leading blank line
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// Strip trailing blank lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return [][]rune{{}}, 0, 0
	}

	maxW := 0
	for _, line := range lines {
		if n := len([]rune(line)); n > maxW {
			maxW = n
		}
	}

	grid := make([][]rune, len(lines))
	for i, line := range lines {
		runes := []rune(line)
		row := make([]rune, maxW)
		for j := range row {
			if j < len(runes) {
				row[j] = runes[j]
			} else {
				row[j] = ' '
			}
		}
		grid[i] = row
	}
	return grid, maxW, len(lines)
}

// ParseColorFrames converts color mask strings into byte grids.
func ParseColorFrames(masks []string) [][][]byte {
	var frames [][][]byte
	for _, s := range masks {
		frames = append(frames, parseColorFrame(s))
	}
	return frames
}

func parseColorFrame(s string) [][]byte {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	maxW := 0
	for _, line := range lines {
		if len(line) > maxW {
			maxW = len(line)
		}
	}
	grid := make([][]byte, len(lines))
	for i, line := range lines {
		row := make([]byte, maxW)
		for j := range row {
			if j < len(line) {
				row[j] = line[j]
			}
		}
		grid[i] = row
	}
	return grid
}

// DetectTransparent picks the transparent char for auto_trans entities.
// If '?' appears anywhere, use '?'. Otherwise use ' '.
func DetectTransparent(frames [][][]rune) rune {
	for _, frame := range frames {
		for _, row := range frame {
			for _, ch := range row {
				if ch == '?' {
					return '?'
				}
			}
		}
	}
	return ' '
}

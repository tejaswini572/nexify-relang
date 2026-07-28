package main

import (
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Engine manages entities, animation, collision detection, and rendering.
type Engine struct {
	screen   tcell.Screen
	entities []*Entity
	width    int
	height   int
}

func NewEngine(screen tcell.Screen) *Engine {
	w, h := screen.Size()
	return &Engine{screen: screen, width: w, height: h}
}

func (eng *Engine) AddEntity(e *Entity)    { eng.entities = append(eng.entities, e) }
func (eng *Engine) RemoveAllEntities()     { eng.entities = nil }
func (eng *Engine) Width() int             { return eng.width }
func (eng *Engine) Height() int            { return eng.height }
func (eng *Engine) UpdateSize()            { eng.width, eng.height = eng.screen.Size() }

// Animate performs one tick: move, animate, collide, clean up.
func (eng *Engine) Animate() {
	now := time.Now().Unix()

	for _, e := range eng.entities {
		if e.Dead {
			continue
		}
		if e.Callback != nil {
			e.Callback(e, eng)
		} else {
			e.X += e.DX
			e.Y += e.DY
		}
		if e.AnimSpeed > 0 && len(e.Frames) > 1 {
			e.AnimFrame += e.AnimSpeed
			e.AnimTotal += e.AnimSpeed
			for e.AnimFrame >= float64(len(e.Frames)) {
				e.AnimFrame -= float64(len(e.Frames))
			}
		}
		if e.DieOffscreen && e.IsOffscreen(eng.width, eng.height) {
			e.Dead = true
		}
		if e.DieTime > 0 && now >= e.DieTime {
			e.Dead = true
		}
		if e.DieFrame > 0 && e.AnimTotal >= float64(e.DieFrame) {
			e.Dead = true
		}
	}

	eng.detectCollisions()

	for _, e := range eng.entities {
		if !e.Dead && e.CollHandler != nil && len(e.Collisions) > 0 {
			e.CollHandler(e, eng)
		}
	}

	eng.cleanup()
}

func (eng *Engine) detectCollisions() {
	for _, e := range eng.entities {
		e.Collisions = nil
	}
	n := len(eng.entities)
	for i := 0; i < n; i++ {
		a := eng.entities[i]
		if a.Dead || !a.Physical {
			continue
		}
		for j := i + 1; j < n; j++ {
			b := eng.entities[j]
			if b.Dead || !b.Physical {
				continue
			}
			// Z-range overlap
			if a.Z > b.Z+b.CollDepth || b.Z > a.Z+a.CollDepth {
				continue
			}
			// 2D bounding box overlap
			ax, ay := a.IntX(), a.IntY()
			bx, by := b.IntX(), b.IntY()
			if ax+a.Width <= bx || bx+b.Width <= ax || ay+a.Height <= by || by+b.Height <= ay {
				continue
			}
			
			// Pixel-perfect overlap check
			overlap := false
			startX := ax
			if bx > startX {
				startX = bx
			}
			endX := ax + a.Width
			if bx+b.Width < endX {
				endX = bx + b.Width
			}
			startY := ay
			if by > startY {
				startY = by
			}
			endY := ay + a.Height
			if by+b.Height < endY {
				endY = by + b.Height
			}

			for y := startY; y < endY; y++ {
				for x := startX; x < endX; x++ {
					if a.IsSolid(x-ax, y-ay) && b.IsSolid(x-bx, y-by) {
						overlap = true
						break
					}
				}
				if overlap {
					break
				}
			}

			if !overlap {
				continue
			}
			a.Collisions = append(a.Collisions, b)
			b.Collisions = append(b.Collisions, a)
		}
	}
}

func (eng *Engine) cleanup() {
	alive := eng.entities[:0]
	var dead []*Entity
	for _, e := range eng.entities {
		if e.Dead {
			dead = append(dead, e)
		} else {
			alive = append(alive, e)
		}
	}
	eng.entities = alive
	for _, e := range dead {
		if e.DeathCB != nil {
			e.DeathCB(e, eng)
		}
	}
}

// Render draws all entities to the screen in z-order (back to front).
func (eng *Engine) Render() {
	eng.screen.Clear()

	sorted := make([]*Entity, len(eng.entities))
	copy(sorted, eng.entities)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Z > sorted[j].Z
	})

	for _, e := range sorted {
		if e.Dead {
			continue
		}
		frame := e.CurrentFrame()
		if frame == nil {
			continue
		}
		cf := e.CurrentColorFrame()
		ix, iy := e.IntX(), e.IntY()

		for row := 0; row < len(frame); row++ {
			sy := iy + row
			if sy < 0 || sy >= eng.height {
				continue
			}
			for col := 0; col < len(frame[row]); col++ {
				sx := ix + col
				if sx < 0 || sx >= eng.width {
					continue
				}
				ch := frame[row][col]
				if e.TransChar != 0 && ch == e.TransChar {
					continue
				}
				style := e.DefStyle
				if cf != nil && row < len(cf) && col < len(cf[row]) {
					if c := cf[row][col]; c != ' ' && c != 0 {
						if s, ok := colorMap[c]; ok {
							style = s
						}
					}
				}
				eng.screen.SetContent(sx, sy, ch, nil, style)
			}
		}
	}
	eng.screen.Show()
}

func (eng *Engine) GetEntitiesOfType(t string) []*Entity {
	var result []*Entity
	for _, e := range eng.entities {
		if !e.Dead && e.Type == t {
			result = append(result, e)
		}
	}
	return result
}

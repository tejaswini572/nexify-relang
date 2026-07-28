package main

import (
	"math/rand"
	"time"
)

type PipesScreen struct {
	config   PipeConfig
	renderer *Renderer
	pipes    []*Pipe
	width    int
	height   int
	count    int
	delay    time.Duration
}

func NewPipesScreen(config PipeConfig, width, height int, renderer *Renderer) *PipesScreen {
	ps := &PipesScreen{
		config:   config,
		renderer: renderer,
		width:    width,
		height:   height,
		delay:    time.Duration(float64(time.Second) / float64(config.Fps)),
	}
	ps.initPipes()
	return ps
}

func (ps *PipesScreen) initPipes() {
	ps.pipes = make([]*Pipe, 0, ps.config.Pipes)
	for i := 0; i < ps.config.Pipes; i++ {
		dir := UP
		x := ps.width / 2
		y := ps.height / 2

		if ps.config.RandomStart {
			dir = Direction(rand.Intn(4))
			x = rand.Intn(ps.width)
			y = rand.Intn(ps.height)
		}

		pipeType := ps.config.PipeTypes[rand.Intn(len(ps.config.PipeTypes))]
		color := ps.config.Colors[rand.Intn(len(ps.config.Colors))]

		ps.pipes = append(ps.pipes, &Pipe{
			X:         x,
			Y:         y,
			Direction: dir,
			PipeType:  pipeType,
			Color:     color,
		})
	}
}

func (ps *PipesScreen) UpdateState(width, height int) {
	if width != ps.width || height != ps.height {
		ps.width = width
		ps.height = height
		ps.renderer.Clear()
	}

	ps.updatePipes()
	
	ps.count += len(ps.pipes)
	if ps.config.Limit > 0 && ps.count >= ps.config.Limit {
		ps.renderer.Clear()
		ps.count = 0
	}
	
	time.Sleep(ps.delay)
}

func (ps *PipesScreen) updatePipes() {
	for _, pipe := range ps.pipes {
		x, y := pipe.X, pipe.Y
		oldDir := pipe.Direction

		if oldDir%2 != 0 {
			x += -int(oldDir) + 2
		} else {
			y += int(oldDir) - 1
		}

		if x < 0 || x >= ps.width || y < 0 || y >= ps.height {
			if !ps.config.KeepStyle {
				pipe.PipeType = ps.config.PipeTypes[rand.Intn(len(ps.config.PipeTypes))]
				pipe.Color = ps.config.Colors[rand.Intn(len(ps.config.Colors))]
			}
			x = (x%ps.width + ps.width) % ps.width
			y = (y%ps.height + ps.height) % ps.height
		}

		newDir := oldDir
		if rand.Intn(ps.config.Steady) <= 1 {
			turn := 2*rand.Intn(2) - 1
			newDir = Direction((int(oldDir) + turn + 4) % 4)
		}

		ps.renderer.DrawPipe(pipe, oldDir, newDir)

		pipe.X = x
		pipe.Y = y
		pipe.Direction = newDir
	}
}

func (ps *PipesScreen) HandleKey(key rune) bool {
	switch key {
	case 'P', 'p':
		if ps.config.Steady < 15 {
			ps.config.Steady++
		}
	case 'O', 'o':
		if ps.config.Steady > 3 {
			ps.config.Steady--
		}
	case 'F', 'f':
		if ps.config.Fps < 100 {
			ps.config.Fps++
			ps.delay = time.Duration(float64(time.Second) / float64(ps.config.Fps))
		}
	case 'D', 'd':
		if ps.config.Fps > 20 {
			ps.config.Fps--
			ps.delay = time.Duration(float64(time.Second) / float64(ps.config.Fps))
		}
	case 'B', 'b':
		ps.config.Bold = !ps.config.Bold
		ps.renderer.SetConfig(ps.config)
	case 'C', 'c':
		ps.config.Color = !ps.config.Color
		ps.renderer.SetConfig(ps.config)
	case 'K', 'k':
		ps.config.KeepStyle = !ps.config.KeepStyle
	case '?', 27: // ESC is ASCII 27
		return false
	}
	return true
}

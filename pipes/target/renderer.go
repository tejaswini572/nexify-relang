package main

import (
	"fmt"
)

var PipeSets = []string{
	"┃┏ ┓┛━┓  ┗┃┛┗ ┏━",
	"│╭ ╮╯─╮  ╰│╯╰ ╭─",
	"│┌ ┐┘─┐  └│┘└ ┌─",
	"║╔ ╗╝═╗  ╚║╝╚ ╔═",
	"|+ ++-+  +|++ +-",
	"|/ \\ /-\\  \\|/\\ /-",
	".o ....  .... .o",
	".o oo.o  o.oo o.",
	"-\\ /\\|/  /-\\/ \\|",
	"╿┍ ┑┚╼┒  ┕╽┙┖ ┎╾",
}

type Renderer struct {
	config PipeConfig
	sets   []rune
}

func NewRenderer(config PipeConfig) *Renderer {
	r := &Renderer{
		config: config,
	}
	
	// Prepare sets, padding to 16 characters each if needed
	for _, ps := range PipeSets {
		runes := []rune(ps)
		for len(runes) < 16 {
			runes = append(runes, ' ')
		}
		r.sets = append(r.sets, runes[:16]...)
	}
	
	r.Clear()
	return r
}

func (r *Renderer) SetConfig(config PipeConfig) {
	r.config = config
}

func (r *Renderer) Clear() {
	// clear screen and move cursor to top left
	fmt.Print("\033[2J\033[H")
}

func (r *Renderer) HideCursor() {
	fmt.Print("\033[?25l")
}

func (r *Renderer) ShowCursor() {
	// show cursor and reset all attributes
	fmt.Print("\033[?25h\033[0m")
}

func (r *Renderer) DrawPipe(pipe *Pipe, oldDir, newDir Direction) {
	base := pipe.PipeType * 16
	index := base + int(oldDir)*4 + int(newDir)
	var char rune = '?'
	if index < len(r.sets) {
		char = r.sets[index]
	}

	var colorStr string
	if !r.config.Color {
		colorStr = "\033[0m"
	} else {
		// curses colors mapped to standard ANSI: 0=black, 1=red, 2=green, 3=yellow, 4=blue, 5=magenta, 6=cyan, 7=white
		c := pipe.Color % 8
		if r.config.Bold {
			colorStr = fmt.Sprintf("\033[1;3%dm", c)
		} else {
			colorStr = fmt.Sprintf("\033[0;3%dm", c)
		}
	}

	// Move cursor to (X, Y) and print. ANSI is 1-indexed.
	fmt.Printf("\033[%d;%dH%s%c", pipe.Y+1, pipe.X+1, colorStr, char)
}

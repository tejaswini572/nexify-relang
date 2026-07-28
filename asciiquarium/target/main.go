package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer screen.Fini()

	screen.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite))
	screen.Clear()

	eng := NewEngine(screen)

	events := make(chan tcell.Event, 32)
	go func() {
		for {
			ev := screen.PollEvent()
			if ev == nil {
				close(events)
				return
			}
			events <- ev
		}
	}()

	// Outer loop: each iteration sets up and runs the scene.
	// 'r' breaks the inner loop to recreate; 'q' quits entirely.
	for {
		eng.UpdateSize()
		eng.RemoveAllEntities()
		addEnvironment(eng)
		addCastle(eng)
		addAllSeaweed(eng)
		addAllFish(eng)
		randomObject(eng)
		eng.Render()

		if !runLoop(eng, screen, events) {
			return
		}
	}
}

// runLoop runs the animation loop. Returns false to quit, true to redraw.
func runLoop(eng *Engine, screen tcell.Screen, events <-chan tcell.Event) bool {
	quit := make(chan struct{})

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	paused := false

	for {
		select {
		case <-quit:
			return false

		case ev, ok := <-events:
			if !ok {
				return false
			}
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlC || ev.Key() == tcell.KeyEscape {
					return false
				}
				r := ev.Rune()
				if r == 'q' || r == 'Q' {
					return false
				}
				if r == 'r' || r == 'R' {
					return true
				}
				if r == 'p' || r == 'P' {
					paused = !paused
				}
			case *tcell.EventResize:
				screen.Sync()
				eng.UpdateSize()
			}

		case <-ticker.C:
			if !paused {
				eng.Animate()
			}
			eng.Render()
		}
	}
}

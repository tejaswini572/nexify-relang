package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
)

var version = "pipes-py v2.0.0 (Go port)"

func main() {
	rand.Seed(time.Now().UnixNano())
	
	pipesPtr := flag.Int("p", 0, "number of pipes")
	fpsPtr := flag.Int("f", 0, "frames per second (20-100)")
	steadyPtr := flag.Int("s", 0, "steadiness (5-15)")
	limitPtr := flag.Int("r", -1, "character limit before reset")
	randomPtr := flag.Bool("R", false, "random start")
	noBoldPtr := flag.Bool("B", false, "disable bold")
	noColorPtr := flag.Bool("C", false, "disable color")
	pipeStylePtr := flag.Int("P", -1, "change pipe style (0-9)")
	keepStylePtr := flag.Bool("K", false, "keep style on wrap")
	saveConfigPtr := flag.Bool("S", false, "save current settings as default")
	versionPtr := flag.Bool("v", false, "show version")
	
	// Add long aliases
	flag.IntVar(pipesPtr, "pipes", 0, "number of pipes")
	flag.IntVar(fpsPtr, "fps", 0, "frames per second (20-100)")
	flag.IntVar(steadyPtr, "steady", 0, "steadiness (5-15)")
	flag.IntVar(limitPtr, "limit", -1, "character limit before reset")
	flag.BoolVar(randomPtr, "random", false, "random start")
	flag.BoolVar(noBoldPtr, "no-bold", false, "disable bold")
	flag.BoolVar(noColorPtr, "no-color", false, "disable color")
	flag.IntVar(pipeStylePtr, "pipe-style", -1, "change pipe style (0-9)")
	flag.BoolVar(keepStylePtr, "keep-style", false, "keep style on wrap")
	flag.BoolVar(saveConfigPtr, "save-config", false, "save current settings as default")
	flag.BoolVar(versionPtr, "version", false, "show version")

	flag.Parse()

	if *versionPtr {
		fmt.Println(version)
		return
	}

	config := LoadConfig()

	if *pipesPtr > 0 {
		config.Pipes = *pipesPtr
	}
	if *fpsPtr > 0 {
		if *fpsPtr < 20 {
			config.Fps = 20
		} else if *fpsPtr > 100 {
			config.Fps = 100
		} else {
			config.Fps = *fpsPtr
		}
	}
	if *steadyPtr > 0 {
		if *steadyPtr < 5 {
			config.Steady = 5
		} else if *steadyPtr > 15 {
			config.Steady = 15
		} else {
			config.Steady = *steadyPtr
		}
	}
	if *limitPtr >= 0 {
		config.Limit = *limitPtr
	}
	if *randomPtr {
		config.RandomStart = true
	}
	if *noBoldPtr {
		config.Bold = false
	}
	if *noColorPtr {
		config.Color = false
	}
	if *keepStylePtr {
		config.KeepStyle = true
	}
	if *pipeStylePtr >= 0 && *pipeStylePtr <= 9 {
		config.PipeTypes = []int{*pipeStylePtr}
	}

	if *saveConfigPtr {
		SaveConfig(config)
	}

	// Setup terminal
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		fmt.Println("Not running in a terminal.")
		return
	}
	
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Printf("Error setting raw mode: %v\n", err)
		return
	}
	
	// Ensure we restore state upon returning
	defer term.Restore(fd, oldState)
	
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	renderer := NewRenderer(config)
	renderer.HideCursor()
	defer renderer.ShowCursor()
	
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}
	
	screen := NewPipesScreen(config, width, height, renderer)
	
	keyChan := make(chan rune)
	go func() {
		b := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(b)
			if err != nil || n == 0 {
				continue
			}
			if b[0] == 3 {
				c <- os.Interrupt
				return
			}
			keyChan <- rune(b[0])
		}
	}()

	for {
		select {
		case <-c:
			return
		case k := <-keyChan:
			if !screen.HandleKey(k) {
				return
			}
		default:
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				screen.UpdateState(w, h)
			}
		}
	}
}

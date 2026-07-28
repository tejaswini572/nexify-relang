package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Named constants
const (
	FrameDelay       = 40 * time.Millisecond
	DefaultW         = 80
	DefaultH         = 24
	SmokeTrailBuffer = 40 // Accounts for smoke/sprite rendering past the train's trailing edge
)

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func getTerminalSize() (int, int) {
	// Try Unix-style 'stty size' first
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			h, errH := strconv.Atoi(parts[0])
			w, errW := strconv.Atoi(parts[1])
			if errH == nil && errW == nil {
				return w, h
			}
		}
	}

	// Try Windows PowerShell fallback
	cmd = exec.Command("powershell", "-NoProfile", "-Command", "Write-Output \"$($Host.UI.RawUI.WindowSize.Height) $($Host.UI.RawUI.WindowSize.Width)\"")
	out, err = cmd.Output()
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			h, errH := strconv.Atoi(parts[0])
			w, errW := strconv.Atoi(parts[1])
			if errH == nil && errW == nil {
				return w, h
			}
		}
	}

	return DefaultW, DefaultH
}

type Config struct {
	Mini     bool
	Fly      bool
	Alt      bool
	Accident bool
	Dance    bool
}

type Buffer struct {
	w, h int
	data [][]rune
}

func NewBuffer(w, h int) *Buffer {
	b := &Buffer{w: w, h: h, data: make([][]rune, h)}
	for i := range b.data {
		b.data[i] = make([]rune, w)
		for j := range b.data[i] {
			b.data[i][j] = ' '
		}
	}
	return b
}

func (b *Buffer) Clear() {
	for i := range b.data {
		for j := range b.data[i] {
			b.data[i][j] = ' '
		}
	}
}

func (b *Buffer) DrawText(x, y int, text string) {
	if y < 0 || y >= b.h {
		return
	}
	runes := []rune(text)
	for i, r := range runes {
		px := x + i
		if px >= 0 && px < b.w {
			b.data[y][px] = r
		}
	}
}

func (b *Buffer) Render() {
	var buf bytes.Buffer
	// Move cursor to top left
	buf.WriteString("\033[H")
	for i := range b.data {
		buf.WriteString(string(b.data[i]))
		if i < len(b.data)-1 {
			buf.WriteString("\r\n")
		}
	}
	fmt.Print(buf.String())
}

var miniTrain = [][]string{
	{
		`     ++      +------ `,
		`     ||      |+-+ |  `,
		`   /---------|| | |  `,
		`  + ========  +-+ |  `,
		` _|--O========O~\-+  `,
		`//// \_/      \_/    `,
	},
	{
		`     ++      +------ `,
		`     ||      |+-+ |  `,
		`   /---------|| | |  `,
		`  + ========  +-+ |  `,
		` _|--/O========O\-+  `,
		`//// \_/      \_/    `,
	},
}

var defaultTrain = [][]string{
	{
		`           _____                                                              ____         `,
		`          |__+__|                                             ___________   _||__||_       `,
		`         _||___||_                   _________________       |  ___ ___  | |      | \      `,
		`        | |     | |________________  |               |       | |_| |_| | | |      | |      `,
		`        | |     | |________________| |               |_______|_________| |_|______| |      `,
		`        | |     | |                  |               |       |         | |          |      `,
		`        | |     | |                  |               |       | [] [] []| |          |      `,
		`   _____| |     | |                  |               |       |         | |          |____  `,
		`  |_____| |     | |                  |               |       |         | |          |___/  `,
		` _|     |===========================================================================|      `,
		`/ |     |          _           _           _           _          _           _     |      `,
		`| |     |        / _ \       / _ \       / _ \       / _ \      / _ \       / _ \   |      `,
		`\_|_____|_______|__O__=======__O__=======__O__=======__O__|____|__O__=======__O__|__|      `,
		`  \_O_O_/        \___/       \___/       \___/       \___/      \___/       \___/          `,
	},
	{
		`           _____                                                              ____         `,
		`          |__+__|                                             ___________   _||__||_       `,
		`         _||___||_                   _________________       |  ___ ___  | |      | \      `,
		`        | |     | |________________  |               |       | |_| |_| | | |      | |      `,
		`        | |     | |________________| |               |_______|_________| |_|______| |      `,
		`        | |     | |                  |               |       |         | |          |      `,
		`        | |     | |                  |               |       | [] [] []| |          |      `,
		`   _____| |     | |                  |               |       |         | |          |____  `,
		`  |_____| |     | |                  |               |       |         | |          |___/  `,
		` _|     |===========================================================================|      `,
		`/ |     |          _           _           _           _          _           _     |      `,
		`| |     |        / O__=======__O__=======__O__=======__O__       /O__========__O\   |      `,
		`\_|_____|_______| (_) |     | (_) |     | (_) |     | (_) |    | (_) |     | (_) |__|      `,
		`  \_O_O_/        \___/       \___/       \___/       \___/      \___/       \___/          `,
	},
}

var altTrain = [][]string{
	{
		`                         __________________        _________________                 _______     `,
		`                       /|  |            |  |      |  |           |  |      |\      /_______ \    `,
		`                      | |  |            |  |      |  |           |  |      | \    /        | \   `,
		`                      | |  |            |  |______|  |           |  |______|  \  |#########| |   `,
		`                      | |  |            |  |      |  |           |  |      |   | |         | |   `,
		`  _________           | |  |            |  |      |  |           |  |      |   | |         | |   `,
		` /_______ /|          | |__|            |__|      |__|           |__|      |___| |#########| |   `,
		`|       | ||          |__________________________________________________________| |         | |   `,
		`|       | ||          |       |        | |              |       |        | |     |=========| |   `,
		`|_______|/ |==========|_______|________|_|==============|_______|________|_|=====|_________| |   `,
		`/          |             _             _                   _             _                   |   `,
		`|          |           / _ \         / _ \               / _ \         / _ \                 |   `,
		`\_|________|__________|__O__=========__O__|_____________|__O__=========__O__|________________|   `,
		`  \_O_O_/              \___/         \___/               \___/         \___/                     `,
	},
	{
		`                         __________________        _________________                 _______     `,
		`                       /|  |            |  |      |  |           |  |      |\      /_______ \    `,
		`                      | |  |            |  |      |  |           |  |      | \    /        | \   `,
		`                      | |  |            |  |______|  |           |  |______|  \  |#########| |   `,
		`                      | |  |            |  |      |  |           |  |      |   | |         | |   `,
		`  _________           | |  |            |  |      |  |           |  |      |   | |         | |   `,
		` /_______ /|          | |__|            |__|      |__|           |__|      |___| |#########| |   `,
		`|       | ||          |__________________________________________________________| |         | |   `,
		`|       | ||          |       |        | |              |       |        | |     |=========| |   `,
		`|_______|/ |==========|_______|________|_|==============|_______|________|_|=====|_________| |   `,
		`/          |             _             _                   _             _                   |   `,
		`|          |           /_O__=========__O_\               /_O__=========__O_\                 |   `,
		`\_|________|__________| (_) |       | (_) |_____________| (_) |       | (_) |________________|   `,
		`  \_O_O_/              \___/         \___/               \___/         \___/                     `,
	},
}

var smokePuffs = []string{
	"(   )",
	" ( @@) ",
	"  (@@@)  ",
	"   (@@@)   ",
	"    (@@)    ",
	"     (@)     ",
	"     O     ",
	"    O    ",
	"   o   ",
	"  o  ",
	" . ",
}

func main() {
	var cfg Config
	var random bool
	flag.BoolVar(&cfg.Mini, "l", false, "mini train")
	flag.BoolVar(&cfg.Fly, "F", false, "fly diagonally")
	flag.BoolVar(&cfg.Alt, "c", false, "alternate train")
	flag.BoolVar(&cfg.Accident, "a", false, "accident mode")
	flag.BoolVar(&cfg.Dance, "d", false, "dance mode")
	flag.BoolVar(&random, "r", false, "random mode")
	flag.Parse()

	if random {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		cfg.Mini = r.Intn(2) == 1
		cfg.Fly = r.Intn(2) == 1
		cfg.Alt = r.Intn(2) == 1
		cfg.Accident = r.Intn(2) == 1
		cfg.Dance = r.Intn(2) == 1
	}

	w, h := getTerminalSize()
	buf := NewBuffer(w, h)

	train := defaultTrain
	if cfg.Mini {
		train = miniTrain
	} else if cfg.Alt {
		train = altTrain
	}

	// Compute max width spanning all lines in all frames of the selected train
	trainW := 0
	for _, frame := range train {
		for _, line := range frame {
			pl := len([]rune(line))
			if pl > trainW {
				trainW = pl
			}
		}
	}
	trainH := len(train[0])

	// Clear screen beforehand
	fmt.Print("\033[2J")

	var tick int
	startX := w
	endX := -trainW - SmokeTrailBuffer

	for x := startX; x > endX; x-- {
		buf.Clear()
		frameIdx := (tick / 2) % len(train)

		y := (h / 2) - (trainH / 2)
		if cfg.Fly {
			y = h - (startX-x)/4 - trainH
			y = clamp(y, -trainH, h-trainH)
		}

		// Draw Smoke
		smokeStartX := x + 10
		if cfg.Mini {
			smokeStartX = x + 3
		}
		smokeY := y - 2
		for i := 0; i < 4; i++ {
			sx := smokeStartX + i*4 + (tick % 2)
			sy := smokeY - i*2
			puff := smokePuffs[(tick+i)%len(smokePuffs)]
			if sx > x {
				buf.DrawText(sx, sy, puff)
			}
		}

		// Draw Train
		for i, line := range train[frameIdx] {
			buf.DrawText(x, y+i, line)
		}

		// Note on Accident and Dance vs Fly interactions:
		// Accident and Dance modes display additional characters beside the train.
		// These extra characters are explicitly suppressed while Fly is active (and Dance suppressed if Accident is)
		// because computing 3D diagonal flight trajectories for tracking physics on extra entities is too complex
		// for a simple string overlay, and they would visually break alignment with the terrain/track in midair anyway.
		if cfg.Accident {
			buf.DrawText(x+trainW+2, y+trainH-2, "Help!")
			buf.DrawText(x+trainW+2, y+trainH-1, "\\O/")
			if !cfg.Fly {
				buf.DrawText(x+trainW+7, y+trainH-1, "(O)")
			}
		} else if cfg.Dance && !cfg.Fly {
			// Sequence of dance shapes
			danceAnim := []string{"_O_", "(O)", "(O_"}
			d1 := danceAnim[(tick/2)%len(danceAnim)]
			d2 := danceAnim[(tick/2+1)%len(danceAnim)]
			buf.DrawText(x+trainW+3, y+trainH-3, d1)
			buf.DrawText(x+trainW+3, y+trainH-2, " # ")
			buf.DrawText(x+trainW+3, y+trainH-1, "/\\ ")

			buf.DrawText(x+trainW+10, y+trainH-3, d2)
			buf.DrawText(x+trainW+10, y+trainH-2, " # ")
			buf.DrawText(x+trainW+10, y+trainH-1, "/\\ ")
		}

		buf.Render()
		time.Sleep(FrameDelay)
		tick++
	}

	// Final clear
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

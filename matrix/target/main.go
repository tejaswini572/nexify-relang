package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var matrixChars = []string{
	"- ", "* ", "% ", "& ", "# ", "@ ", "1 ", "2 ", "3 ", "4 ", "5 ", "6 ", "7 ", "8 ", "9 ", "0 ",
	"ア", "ィ", "イ", "ゥ", "ウ", "ェ", "エ", "ォ", "オ", "カ", "ガ", "キ", "ギ", "ク", "グ", "ケ", "ゲ", "コ",
	"ゴ", "サ", "ザ", "シ", "ジ", "ス", "ズ", "セ", "ゼ", "ソ", "ゾ", "タ", "ダ", "チ", "ヂ", "ッ", "ツ", "ヅ", "テ",
}

var terminalColours = []string{"22", "28"}

type Matrix struct {
	screenWidth int
	lineCount   int
	lineSpeed   float64
	rng         *rand.Rand
}

func (m *Matrix) getCharacterString() string {
	return matrixChars[m.rng.Intn(len(matrixChars))]
}

func (m *Matrix) getTextColourRandomChar() string {
	return "\033[38;5;" + terminalColours[m.rng.Intn(len(terminalColours))] + "m"
}

func (m *Matrix) getTextColourLightGreenChar() string {
	return "\033[38;5;15m"
}

func main() {
	screenWidth := flag.Int("width", 150, "can be used to change the width of the output in the terminal")
	lineCount := flag.Int("lines", 750, "can be used to change the amount of lines")
	lineSpeed := flag.Float64("speed", 0.1, "how quickly the lines print")
	flag.Parse()

	if *screenWidth < 1 {
		fmt.Fprintln(os.Stderr, "error: -width must be a positive integer")
		os.Exit(1)
	}
	if *lineCount < 1 {
		fmt.Fprintln(os.Stderr, "error: -lines must be a positive integer")
		os.Exit(1)
	}
	if *lineSpeed <= 0 {
		fmt.Fprintln(os.Stderr, "error: -speed must be greater than 0")
		os.Exit(1)
	}

	m := &Matrix{
		screenWidth: *screenWidth,
		lineCount:   *lineCount,
		lineSpeed:   *lineSpeed,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// lineArray holds the state for each column.
	// 0: empty
	// 1: trailing char (normal color)
	// 2: leading char (bright color)
	lineArray := make([]int, m.screenWidth)
	for i := range lineArray {
		lineArray[i] = 1
	}

	// Reset ANSI attributes on exit
	defer fmt.Print("\033[0m")

	// Catch SIGINT to ensure ANSI reset happens on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print("\033[0m\n")
		os.Exit(0)
	}()

	for l := 0; l < m.lineCount; l++ {
		var line strings.Builder

		for col, n := range lineArray {
			if n == 1 || n == 2 {
				if n == 2 {
					line.WriteString(m.getTextColourLightGreenChar())
					line.WriteString(m.getCharacterString())
					lineArray[col] = 1
				} else {
					line.WriteString(m.getTextColourRandomChar())
					line.WriteString(m.getCharacterString())
				}

				if m.rng.Intn(30)+1 == 1 {
					lineArray[col] = 0
				}
			} else {
				line.WriteString(m.getTextColourRandomChar())
				line.WriteString(" ")
				if m.rng.Intn(60)+1 == 1 {
					lineArray[col] = 2
				}
			}
		}

		fmt.Println(line.String())
		time.Sleep(time.Duration(m.lineSpeed * float64(time.Second)))
	}
}

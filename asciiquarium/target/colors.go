package main

import (
	"bytes"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// colorMap maps single-character color codes to tcell styles.
// Lowercase = normal, Uppercase = bold (bright).
var colorMap = map[byte]tcell.Style{
	'k': tcell.StyleDefault.Foreground(tcell.ColorBlack),
	'K': tcell.StyleDefault.Foreground(tcell.ColorBlack).Bold(true),
	'r': tcell.StyleDefault.Foreground(tcell.ColorRed),
	'R': tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true),
	'g': tcell.StyleDefault.Foreground(tcell.ColorGreen),
	'G': tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true),
	'y': tcell.StyleDefault.Foreground(tcell.ColorYellow),
	'Y': tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
	'b': tcell.StyleDefault.Foreground(tcell.ColorBlue),
	'B': tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true),
	'm': tcell.StyleDefault.Foreground(tcell.ColorFuchsia),
	'M': tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	'c': tcell.StyleDefault.Foreground(tcell.ColorAqua),
	'C': tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true),
	'w': tcell.StyleDefault.Foreground(tcell.ColorWhite),
	'W': tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),
}

// namedStyle converts a named color (e.g. "cyan", "CYAN", "RED") to a tcell style.
func namedStyle(name string) tcell.Style {
	lower := strings.ToLower(name)
	bold := name != lower

	var fg tcell.Color
	switch lower {
	case "black":
		fg = tcell.ColorBlack
	case "red":
		fg = tcell.ColorRed
	case "green":
		fg = tcell.ColorGreen
	case "yellow":
		fg = tcell.ColorYellow
	case "blue":
		fg = tcell.ColorBlue
	case "magenta":
		fg = tcell.ColorFuchsia
	case "cyan":
		fg = tcell.ColorAqua
	case "white":
		fg = tcell.ColorWhite
	default:
		fg = tcell.ColorSilver
	}

	s := tcell.StyleDefault.Foreground(fg)
	if bold {
		s = s.Bold(true)
	}
	return s
}

// randColor replaces digit characters 1-9 in a color mask string with random
// color codes. Digit '4' should be pre-replaced with 'W' before calling.
func randColor(mask string) string {
	// 12 colors
	colors := []byte{'c', 'C', 'r', 'R', 'y', 'Y', 'b', 'B', 'g', 'G', 'm', 'M'}
	result := []byte(mask)
	for digit := byte('1'); digit <= byte('9'); digit++ {
		c := colors[rand.Intn(len(colors))]
		result = bytes.ReplaceAll(result, []byte{digit}, []byte{c})
	}
	return string(result)
}

// replaceBackticks replaces § with backtick — workaround for Go raw strings not allowing backticks.
func replaceBackticks(s string) string {
	return strings.ReplaceAll(s, "§", "`")
}

package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// DrawBubble creates the speech or thought bubble around the provided lines of text.
func DrawBubble(lines []string, think bool) string {
	// Empty message edge case
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return " __\n --"
	}

	// Calculate maximum line width
	maxWidth := 0
	for _, line := range lines {
		count := utf8.RuneCountInString(line)
		if count > maxWidth {
			maxWidth = count
		}
	}

	var sb strings.Builder
	
	// Top border
	sb.WriteString(" " + strings.Repeat("_", maxWidth+2) + "\n")

	// Lines
	for i, line := range lines {
		// Calculate right padding
		padLen := maxWidth - utf8.RuneCountInString(line)
		paddedLine := line + strings.Repeat(" ", padLen)

		var left, right string
		if think {
			left, right = "(", ")"
		} else {
			if len(lines) == 1 {
				left, right = "<", ">"
			} else if i == 0 {
				left, right = "/", "\\"
			} else if i == len(lines)-1 {
				left, right = "\\", "/"
			} else {
				left, right = "|", "|"
			}
		}

		sb.WriteString(fmt.Sprintf("%s %s %s\n", left, paddedLine, right))
	}

	// Bottom border
	sb.WriteString(" " + strings.Repeat("-", maxWidth+2))

	return sb.String()
}

func main() {
	fmt.Println("--- STEP 4: BUBBLE DRAWING TESTS ---")

	testCases := []struct {
		name  string
		lines []string
	}{
		{
			name:  "0 lines (Empty Message)",
			lines: []string{""},
		},
		{
			name:  "1 line",
			lines: []string{"Hello, World!"},
		},
		{
			name:  "3 lines",
			lines: []string{"This is a", "multiline message", "test"},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n[%s - Say Mode]\n", tc.name)
		fmt.Println(DrawBubble(tc.lines, false))

		fmt.Printf("\n[%s - Think Mode]\n", tc.name)
		fmt.Println(DrawBubble(tc.lines, true))
	}
}

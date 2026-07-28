package main

import (
	"fmt"
	"strings"
)

// Wrap breaks a message into lines based on the specified column width.
// If width is <= 0, no word wrapping is applied (only explicit newlines are respected).
func Wrap(message string, width int) []string {
	// If no-wrap mode
	if width <= 0 {
		return strings.Split(message, "\n")
	}

	var result []string
	
	// Preserve explicit newlines
	paragraphs := strings.Split(message, "\n")
	for _, paragraph := range paragraphs {
		words := strings.Fields(paragraph) // Collapses whitespace
		
		if len(words) == 0 {
			result = append(result, "")
			continue
		}

		var currentLine string
		
		for _, word := range words {
			// Handle overlong single words
			if len(word) > width {
				if currentLine != "" {
					result = append(result, currentLine)
					currentLine = ""
				}
				
				// Hard-wrap the overlong word exactly at width
				for len(word) > width {
					result = append(result, word[:width])
					word = word[width:]
				}
				if len(word) > 0 {
					currentLine = word
				}
				continue
			}
			
			if currentLine == "" {
				currentLine = word
			} else if len(currentLine)+1+len(word) > width {
				result = append(result, currentLine)
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
		
		if currentLine != "" {
			result = append(result, currentLine)
		}
	}
	
	return result
}

func main() {
	fmt.Println("--- STEP 3: WORD-WRAP TESTS ---")

	testCases := []struct {
		name    string
		message string
		width   int
	}{
		{
			name:    "1. Standard Wrap (width 10)",
			message: "This is a simple wrapping test",
			width:   10,
		},
		{
			name:    "2. Overlong Single Word (width 5)",
			message: "aaaaaaaaaaaaaaaaaa", // 18 'a's
			width:   5,
		},
		{
			name:    "3. Existing Newlines",
			message: "hello\nworld\nthis is long",
			width:   10,
		},
		{
			name:    "4. Whitespace Collapsing",
			message: "  lots   of   space   ",
			width:   10,
		},
		{
			name:    "5. No Wrap Mode (width 0)",
			message: "This is a very long line that should not be wrapped at all\nBut explicit newlines still work",
			width:   0,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n[%s]\n", tc.name)
		wrapped := Wrap(tc.message, tc.width)
		for i, line := range wrapped {
			fmt.Printf("Line %d: '%s'\n", i+1, line)
		}
	}
}

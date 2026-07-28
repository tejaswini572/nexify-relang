package main

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

//go:embed cows/*.cow
var cowsFS embed.FS

// --- STEP 2: Cow Loading ---
func LoadCow(name string) (string, error) {
	data, err := cowsFS.ReadFile("cows/" + name + ".cow")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	var processed []string
	inTemplate := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inTemplate {
			if strings.HasPrefix(trimmed, "$the_cow =") && strings.Contains(trimmed, "EOC") {
				inTemplate = true
			}
			continue
		}
		if trimmed == "EOC" {
			break
		}
		
		line = strings.ReplaceAll(line, "\\\\", "\\")
		line = strings.ReplaceAll(line, "\\@", "@")
		processed = append(processed, line)
	}

	return strings.Join(processed, "\n"), nil
}

// --- STEP 3: Word Wrap ---
func Wrap(message string, width int) []string {
	if width <= 0 {
		return strings.Split(message, "\n")
	}

	var result []string
	paragraphs := strings.Split(message, "\n")
	for _, paragraph := range paragraphs {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}

		var currentLine string
		for _, word := range words {
			if len(word) > width {
				if currentLine != "" {
					result = append(result, currentLine)
					currentLine = ""
				}
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

// --- STEP 4: Bubble Drawing ---
func DrawBubble(lines []string, think bool) string {
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return " __\n --"
	}

	maxWidth := 0
	for _, line := range lines {
		count := utf8.RuneCountInString(line)
		if count > maxWidth {
			maxWidth = count
		}
	}

	var sb strings.Builder
	sb.WriteString(" " + strings.Repeat("_", maxWidth+2) + "\n")

	for i, line := range lines {
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

	sb.WriteString(" " + strings.Repeat("-", maxWidth+2))
	return sb.String()
}

// --- STEP 5: Substitution & Rendering ---
func Render(message string, width int, cowName string, eyes string, tongue string, think bool) (string, error) {
	template, err := LoadCow(cowName)
	if err != nil {
		return "", err
	}

	// 1. Create the Bubble
	lines := Wrap(message, width)
	bubble := DrawBubble(lines, think)

	// 2. Perform Substitutions on the Cow Template
	connector := "\\"
	if think {
		connector = "o"
	}

	// Make sure placeholders are replaced exactly
	cowArt := strings.ReplaceAll(template, "$thoughts", connector)
	cowArt = strings.ReplaceAll(cowArt, "$eyes", eyes)
	cowArt = strings.ReplaceAll(cowArt, "$tongue", tongue)

	// 3. Combine them
	return bubble + "\n" + cowArt, nil
}

func main() {
	fmt.Println("--- STEP 5: SUBSTITUTION & FULL RENDERING TEST ---")

	// Multi-line message
	msg := "This is a test of the fully integrated cowsay rendering process, combining wrap, bubble, and substitution!"
	
	// Default cow, eyes: **, tongue: U , think: false
	output, err := Render(msg, 40, "default", "**", "U ", false)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println(output)
}

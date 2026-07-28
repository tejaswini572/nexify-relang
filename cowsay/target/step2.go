package main

import (
	"embed"
	"fmt"
	"os"
	"strings"
)

//go:embed cows/*.cow
var cowsFS embed.FS

// Load fetches the cow template by name and strips the Perl heredoc wrappers.
func Load(name string) (string, error) {
	data, err := cowsFS.ReadFile("cows/" + name + ".cow")
	if err != nil {
		return "", err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	
	var processed []string
	inTemplate := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if !inTemplate {
			// Find the start of the heredoc
			if strings.HasPrefix(trimmed, "$the_cow =") && strings.Contains(trimmed, "EOC") {
				inTemplate = true
			}
			continue
		}
		
		// End of heredoc
		if trimmed == "EOC" {
			break
		}
		
		// Perl heredocs interpret backslashes. Standard cowsay escapes `\` and `@`.
		// We'll unescape `\\` to `\` and `\@` to `@`
		line = strings.ReplaceAll(line, "\\\\", "\\")
		line = strings.ReplaceAll(line, "\\@", "@")
		
		processed = append(processed, line)
	}

	return strings.Join(processed, "\n"), nil
}

func main() {
	template, err := Load("default")
	if err != nil {
		fmt.Printf("Error loading default.cow: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("--- DEFAULT COW RAW TEMPLATE ---")
	fmt.Println(template)
	fmt.Println("--------------------------------")
}

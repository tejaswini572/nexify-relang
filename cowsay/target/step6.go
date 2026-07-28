package main

import (
	"embed"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

//go:embed cows/*.cow
var cowsFS embed.FS

// LoadCow fetches the cow template by name and strips the Perl heredoc wrappers.
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

func Render(message string, width int, cowName string, eyes string, tongue string, think bool) (string, error) {
	template, err := LoadCow(cowName)
	if err != nil {
		return "", err
	}

	lines := Wrap(message, width)
	bubble := DrawBubble(lines, think)

	connector := "\\"
	if think {
		connector = "o"
	}

	cowArt := strings.ReplaceAll(template, "$thoughts", connector)
	cowArt = strings.ReplaceAll(cowArt, "$eyes", eyes)
	cowArt = strings.ReplaceAll(cowArt, "$tongue", tongue)

	return bubble + "\n" + cowArt, nil
}

// --- STEP 6: CLI PARSING ---
type Options struct {
	Eyes    string
	Tongue  string
	Width   int
	Cowfile string
	NoWrap  bool
	Random  bool
	List    bool
	Think   bool
	Message string
}

func ParseArgs(args []string) Options {
	opts := Options{
		Eyes:    "oo",
		Tongue:  "  ",
		Width:   40,
		Cowfile: "default",
	}

	var messageArgs []string
	moods := make(map[rune]bool)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--think" {
			opts.Think = true
			continue
		}
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Value flags
			if arg == "-e" || arg == "-T" || arg == "-W" || arg == "-f" {
				if i+1 < len(args) {
					val := args[i+1]
					i++
					switch arg {
					case "-e":
						// Cowsay takes the first two chars of the eye string
						if len(val) > 2 {
							val = val[:2]
						} else if len(val) == 1 {
							val = val + " "
						}
						opts.Eyes = val
					case "-T":
						if len(val) > 2 {
							val = val[:2]
						} else if len(val) == 1 {
							val = val + " "
						}
						opts.Tongue = val
					case "-W":
						fmt.Sscanf(val, "%d", &opts.Width)
					case "-f":
						opts.Cowfile = val
					}
				}
				continue
			}
			
			// Grouped boolean flags (e.g. -bd)
			for _, ch := range arg[1:] {
				switch ch {
				case 'n':
					opts.NoWrap = true
				case 'r':
					opts.Random = true
				case 'l':
					opts.List = true
				case 'b', 'd', 'g', 'p', 's', 't', 'w', 'y':
					moods[ch] = true
				}
			}
			continue
		}

		// Positional argument (part of message)
		messageArgs = append(messageArgs, arg)
	}

	opts.Message = strings.Join(messageArgs, " ")

	// Mood precedence (alphabetical)
	moodList := []rune{'b', 'd', 'g', 'p', 's', 't', 'w', 'y'}
	for _, m := range moodList {
		if moods[m] {
			switch m {
			case 'b':
				opts.Eyes, opts.Tongue = "==", "  "
			case 'd':
				opts.Eyes, opts.Tongue = "xx", "U "
			case 'g':
				opts.Eyes, opts.Tongue = "$$", "  "
			case 'p':
				opts.Eyes, opts.Tongue = "@@", "  "
			case 's':
				opts.Eyes, opts.Tongue = "**", "U "
			case 't':
				opts.Eyes, opts.Tongue = "--", "  "
			case 'w':
				opts.Eyes, opts.Tongue = "OO", "  "
			case 'y':
				opts.Eyes, opts.Tongue = "..", "  "
			}
			break // Precedence applied
		}
	}

	if opts.NoWrap {
		opts.Width = 0
	}

	return opts
}

func RunCLI(args []string) string {
	opts := ParseArgs(args)

	if opts.List {
		entries, _ := cowsFS.ReadDir("cows")
		var names []string
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".cow") {
				names = append(names, strings.TrimSuffix(entry.Name(), ".cow"))
			}
		}
		sort.Strings(names)
		return strings.Join(names, "  ")
	}

	// For STEP 6 we aren't hooking up Stdin yet (that's for later or standard behavior)
	// We'll just run Render on the parsed message.
	output, err := Render(opts.Message, opts.Width, opts.Cowfile, opts.Eyes, opts.Tongue, opts.Think)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return output
}

func main() {
	fmt.Println("--- STEP 6: CLI PARSING TESTS ---")

	testCases := []struct {
		desc string
		args []string
	}{
		{
			"1. Normal invocation",
			[]string{"Hello", "CLI!"},
		},
		{
			"2. Mood flag overriding custom eyes and tongue",
			[]string{"-e", "^^", "-T", "V ", "-d", "I am dead, overrides ^^"},
		},
		{
			"3. Think mode with grouped boolean flags (-bn)",
			[]string{"--think", "-bn", "I am a borg and I do not wrap this extremely long string"},
		},
		{
			"4. List mode",
			[]string{"-l"},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n>>> Test: %s\n", tc.desc)
		fmt.Printf("$ cowsay %s\n", strings.Join(tc.args, " "))
		fmt.Println(RunCLI(tc.args))
	}
}

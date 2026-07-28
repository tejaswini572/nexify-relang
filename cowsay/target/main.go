package main

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

//go:embed cows/*.cow
var cowsFS embed.FS

// --- Help Text ---
const helpText = `
Usage: cli.js [-e eye_string] [-f cowfile] [-h] [-l] [-n] [-T tongue_string] [-W
column] [-bdgpstwy] text

If any command-line arguments are left over after all switches have been
processed, they become the cow's message.

If the program is invoked as cowthink then the cow will think its message
instead of saying it.


Options:
  --version   Show version number                                      [boolean]
  -e          Select the appearance of the cow's eyes.           [default: "oo"]
  -T          The tongue is configurable similarly to the eyes through -T and
              tongue_string.                                     [default: "  "]
  -W          Specifies roughly where the message should be wrapped. The default
              is equivalent to -W 40 i.e. wrap words at or before the 40th
              column.                                     [number] [default: 40]
  -f          Specifies a cow picture file (''cowfile'') to use. It can be
              either a path to a cow file or the name of one of cows included in
              the package.                                  [default: "default"]
  --think     Think the message instead of saying it aloud.            [boolean]
  -b          Mode: Borg                                               [boolean]
  -d          Mode: Dead                                               [boolean]
  -g          Mode: Greedy                                             [boolean]
  -p          Mode: Paranoia                                           [boolean]
  -s          Mode: Stoned                                             [boolean]
  -t          Mode: Tired                                              [boolean]
  -w          Mode: Wired                                              [boolean]
  -y          Mode: Youthful                                           [boolean]
  -h, --help  Show help                                                [boolean]
  -n          If it is specified, the given message will not be word-wrapped.
                                                                       [boolean]
  -r          Select a random cow                                      [boolean]
  -l          List all cowfiles included in this package.              [boolean]
`

// --- Cow Loading ---
func LoadCow(name string) (string, error) {
	data, err := cowsFS.ReadFile("cows/" + name + ".cow")
	if err != nil {
		// Mimic the exact NodeJS ENOENT error required by the spec
		return "", fmt.Errorf("ENOENT: no such file or directory, open '<path_to_node_modules>/cowsay/cows/%s.cow'", name)
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

// --- Logic ---
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

// --- CLI Parsing ---
type Options struct {
	Eyes    string
	Tongue  string
	Width   int
	Cowfile string
	NoWrap  bool
	Random  bool
	List    bool
	Think   bool
	Help    bool
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

	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "--think" {
			opts.Think = true
			continue
		}
		if arg == "--help" || arg == "-h" {
			opts.Help = true
			continue
		}
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			if arg == "-e" || arg == "-T" || arg == "-W" || arg == "-f" {
				if i+1 < len(args) {
					val := args[i+1]
					i++
					switch arg {
					case "-e":
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

			for _, ch := range arg[1:] {
				switch ch {
				case 'n':
					opts.NoWrap = true
				case 'r':
					opts.Random = true
				case 'l':
					opts.List = true
				case 'h':
					opts.Help = true
				case 'b', 'd', 'g', 'p', 's', 't', 'w', 'y':
					moods[ch] = true
				}
			}
			continue
		}

		messageArgs = append(messageArgs, arg)
	}

	opts.Message = strings.Join(messageArgs, " ")

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
			break
		}
	}

	if opts.NoWrap {
		opts.Width = 0
	}

	return opts
}

func main() {
	if len(os.Args) > 0 && (strings.HasSuffix(os.Args[0], "cowthink") || strings.HasSuffix(os.Args[0], "cowthink.exe")) {
		os.Args = append(os.Args, "--think")
	}

	opts := ParseArgs(os.Args)

	if opts.Help {
		fmt.Print(strings.TrimSpace(helpText))
		os.Exit(0)
	}

	if opts.List {
		entries, _ := cowsFS.ReadDir("cows")
		var names []string
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".cow") {
				names = append(names, strings.TrimSuffix(entry.Name(), ".cow"))
			}
		}
		sort.Strings(names)
		fmt.Println(strings.Join(names, "  "))
		os.Exit(0)
	}

	if opts.Random {
		entries, _ := cowsFS.ReadDir("cows")
		var names []string
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".cow") {
				names = append(names, strings.TrimSuffix(entry.Name(), ".cow"))
			}
		}
		rand.Seed(time.Now().UnixNano())
		opts.Cowfile = names[rand.Intn(len(names))]
	}

	// Message Source Precedence
	if opts.Message == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			reader := bufio.NewReader(os.Stdin)
			var stdinBytes []byte
			for {
				b, err := reader.ReadByte()
				if err != nil {
					if err == io.EOF {
						break
					}
					break
				}
				stdinBytes = append(stdinBytes, b)
			}
			// Strip final newline if present (similar to strip-final-newline)
			str := string(stdinBytes)
			if strings.HasSuffix(str, "\n") {
				str = str[:len(str)-1]
			}
			if strings.HasSuffix(str, "\r") {
				str = str[:len(str)-1]
			}
			opts.Message = str
		} else {
			// No positional args and no stdin -> show help
			fmt.Print(strings.TrimSpace(helpText))
			os.Exit(0)
		}
	}

	output, err := Render(opts.Message, opts.Width, opts.Cowfile, opts.Eyes, opts.Tongue, opts.Think)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}

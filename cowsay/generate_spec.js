const fs = require('fs');

let spec = `# cowsay Specification

## 1. CLI Flags and exact semantics
- **Mood Flags (with precedence if combined):**
  - \`-b\` (Borg): eyes \`==\`, tongue \`  \`
  - \`-d\` (Dead): eyes \`xx\`, tongue \`U \`
  - \`-g\` (Greedy): eyes \`$$\`, tongue \`  \`
  - \`-p\` (Paranoia): eyes \`@@\`, tongue \`  \`
  - \`-s\` (Stoned): eyes \`**\`, tongue \`U \`
  - \`-t\` (Tired): eyes \`--\`, tongue \`  \`
  - \`-w\` (Wired): eyes \`OO\`, tongue \`  \`
  - \`-y\` (Youthful): eyes \`..\`, tongue \`  \`
  - *Precedence:* Mood flags are resolved strictly in alphabetical order of their defining character (\`b, d, g, p, s, t, w, y\`). The first matching flag forces the eyes and tongue.
- **Custom Overrides:**
  - \`-e\` overrides the default eyes (\`oo\`).
  - \`-T\` overrides the default tongue (\`  \`).
  - *Resolution Order:* Mood flags completely override custom \`-e\` and \`-T\` flags. If any mood flag is present, its defined eyes and tongue take precedence over anything passed via \`-e\` or \`-T\`.
- **Cowfile Selection (\`-f\`):**
  - \`-f <name>\` selects a cow template. Defaults to \`default\`.
- **Random Cow (\`-r\`):**
  - Randomly selects a cowfile from the bundled cows list.
- **List Mode (\`-l\`):**
  - Prints all available bundled cow files separated exactly by two spaces (\`  \`), in alphabetical order, then exits.
- **Wrap-Width (\`-W\`):**
  - Defaults to \`40\`.
  - Determines the maximum width of the speech bubble text.
  - No environment variables are consulted for this width.
- **No-Wrap (\`-n\`):**
  - Disables word-wrapping entirely.
  - Structurally, it passes \`null\` instead of a number for the width, causing lines to span indefinitely until an explicit newline is encountered.
- **Say vs Think Mode:**
  - Determined by the \`--think\` flag, or by invoking the program as \`cowthink\` (based on the executable name).

## 2. Message Source Precedence
- **CLI Positional Args vs Stdin:** Positional CLI arguments (any unparsed trailing strings) take strict precedence. 
- **Conflict Resolution:** If both positional args and piped stdin are provided, positional arguments are joined with spaces to form the message, and stdin is entirely ignored.

## 3. Default Invocation
- **Conditions:** Program is run with no message argument and no piped stdin (e.g. empty TTY).
- **Placeholder Message:** It does NOT print a placeholder message.
- **Output:** It prints the exact CLI help text showing usage and options.
- **Borders:** There are NO top or bottom border characters printed because no bubble is drawn.
- **Exit Code:** Exits with code \`0\`.

## 4. Word-Wrap Algorithm
- **Overlong Single Words:** A single word that exceeds the column limit is hard-wrapped exactly at the column limit (broken into multiple chunks).
- **Existing Newlines:** Any existing newlines in the input string are preserved and force a line break, restarting the column count for the next line.
- **Whitespace Collapsing:** Whitespace is collapsed around word boundaries when wrapping occurs.

## 5. Bubble-Drawing Algorithm
- **Say Mode (\`cowsay\`):**
  - **Single-line borders:** Left: \`<\`, Right: \`>\`.
  - **Multi-line borders:** 
    - First line: Left: \`/\`, Right: \`\\\`.
    - Middle lines: Left: \`|\`, Right: \`|\`.
    - Last line: Left: \`\\\`, Right: \`/\`.
  - **Connector:** \`\\\`
- **Think Mode (\`cowthink\`):**
  - **Single-line & Multi-line borders:** Left: \`(\`, Right: \`)\` on all lines.
  - **Connector:** \`o\`
- **Common Border Rules:**
  - **Top Border:** A space followed by underscores (\`_\`). The number of underscores is equal to the longest wrapped line length plus 2.
  - **Bottom Border:** A space followed by dashes (\`-\`). The number of dashes is equal to the longest wrapped line length plus 2.
- **Zero-Length Message (\`""\`):**
  - The top border is a space and two underscores (\` __\`).
  - The bottom border is a space and two dashes (\` --\`).
  - **No side borders** are drawn at all.

## 6. Error Handling
- **Unknown Cowfile:** Prints \`Error: ENOENT: no such file or directory, open '<path>'\` (plus a stack trace). Exit code \`1\`.
- **Empty Message Passed (\`""\`):** Prints a 0-width empty bubble (\` __\` and \` --\` with no side borders) and the selected cow. Exit code \`0\`.
- **Malformed Flags:** Unknown flags (e.g., \`--invalid-flag\`) are parsed as boolean or string options by the CLI parser, effectively consuming the following positional arguments. If it absorbs the intended message, the program behaves as if no message was provided, resulting in printing the help text (exit code 0).

## 7. Nondeterminism
- **Random Cow (\`-r\`):** Seeded by the internal JavaScript engine state (\`Math.random()\`). It is not explicitly seeded and thus is NOT reproducible across identical runs.

## 8. Cow Template Format
- **Placeholder Tokens:**
  - \`$thoughts\`: Replaced by the connector character (\`\\\` for say, \`o\` for think).
  - \`$eyes\`: Replaced by the current eye string (e.g., \`oo\`, \`==\`).
  - \`$tongue\`: Replaced by the current tongue string (e.g., \`  \`, \`U \`).
- **Bundled Cowfiles Reference Data:**
`;

const cowsDir = 'source/cows/';
const files = fs.readdirSync(cowsDir).filter(f => f.endsWith('.cow'));
for (const file of files) {
  spec += `\n### ${file.replace('.cow', '')}\n`;
  spec += "\`\`\`perl\n";
  spec += fs.readFileSync(cowsDir + file, 'utf8');
  spec += "\n\`\`\`\n";
}

fs.mkdirSync('spec', { recursive: true });
fs.writeFileSync('spec/SPEC.md', spec);

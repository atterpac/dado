package inline

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// SelectOption displays an interactive selection prompt.
// Returns the index of the selected option.
func SelectOption(prompt string, options []string, descriptions []string) int {
	// Check if we have a TTY
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return selectOptionSimple(prompt, options, descriptions)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return selectOptionSimple(prompt, options, descriptions)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	selected := 0

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	for {
		// Clear and redraw
		fmt.Printf("\r\033[K")
		fmt.Printf("\033[%dA", len(options)+1)

		// Print prompt
		fmt.Printf("\r\033[K  %s%s?%s %s%s%s\n", Bold, Magenta, Reset, Bold, prompt, Reset)

		// Print options
		for i, opt := range options {
			fmt.Print("\r\033[K")
			if i == selected {
				fmt.Printf("  %s%s▸%s %s%s%s", Bold, Cyan, Reset, Cyan, opt, Reset)
				if i < len(descriptions) && descriptions[i] != "" {
					fmt.Printf(" %s%s%s", Dim, descriptions[i], Reset)
				}
			} else {
				fmt.Printf("    %s%s%s", White, opt, Reset)
				if i < len(descriptions) && descriptions[i] != "" {
					fmt.Printf(" %s%s%s", Dim, descriptions[i], Reset)
				}
			}
			fmt.Println()
		}

		// Read key
		buf := make([]byte, 3)
		n, _ := os.Stdin.Read(buf)

		if n == 1 {
			switch buf[0] {
			case 13, 10: // Enter
				return selected
			case 'j', 'J': // vim down
				selected = (selected + 1) % len(options)
			case 'k', 'K': // vim up
				selected = (selected - 1 + len(options)) % len(options)
			case 'q', 3: // q or Ctrl+C
				fmt.Print("\033[?25h")
				term.Restore(int(os.Stdin.Fd()), oldState)
				os.Exit(0)
			case '1', '2', '3', '4', '5':
				idx := int(buf[0] - '1')
				if idx < len(options) {
					return idx
				}
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 65: // Up arrow
				selected = (selected - 1 + len(options)) % len(options)
			case 66: // Down arrow
				selected = (selected + 1) % len(options)
			}
		}
	}
}

func selectOptionSimple(prompt string, options []string, descriptions []string) int {
	fmt.Printf("\n  %s%s?%s %s%s%s\n", Bold, Magenta, Reset, Bold, prompt, Reset)
	for i, opt := range options {
		desc := ""
		if i < len(descriptions) {
			desc = descriptions[i]
		}
		fmt.Printf("    %s%d.%s %s", Dim, i+1, Reset, opt)
		if desc != "" {
			fmt.Printf(" %s%s%s", Dim, desc, Reset)
		}
		fmt.Println()
	}
	fmt.Printf("\n  %sChoice [1]:%s ", Dim, Reset)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	for i := range options {
		if input == fmt.Sprintf("%d", i+1) {
			return i
		}
	}
	return 0
}

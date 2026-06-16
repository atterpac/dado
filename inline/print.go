package inline

import (
	"fmt"
	"strings"
)

const Version = "0.1.0"

var Logo = `
     ╦╦╔═╗
     ║║║ ╦
    ╚╝╩╚═╝`

// PrintLogo prints the dado logo.
func PrintLogo() {
	fmt.Fprintf(out, "%s%s%s%s\n", Bold, Magenta, Logo, Reset)
}

// PrintVersion prints the version with logo.
func PrintVersion() {
	PrintLogo()
	fmt.Fprintf(out, "  %s%sv%s%s\n\n", Dim, White, Version, Reset)
}

// PrintSuccess prints a success message.
func PrintSuccess(msg string) {
	fmt.Fprintf(out, "  %s%s✓%s %s\n", Bold, Green, Reset, msg)
}

// PrintError prints an error message.
func PrintError(msg string) {
	fmt.Fprintf(out, "  %s%s✗%s %s%s%s\n", Bold, Red, Reset, Red, msg, Reset)
}

// PrintInfo prints an info message.
func PrintInfo(msg string) {
	fmt.Fprintf(out, "  %s%s→%s %s\n", Bold, Blue, Reset, msg)
}

// PrintStep prints a step message.
func PrintStep(msg string) {
	fmt.Fprintf(out, "  %s%s•%s %s\n", Dim, White, Reset, msg)
}

// PrintFileCreated prints a file creation message.
func PrintFileCreated(path string) {
	fmt.Fprintf(out, "    %s+%s %s%s%s\n", Green, Reset, Dim, path, Reset)
}

// PrintCommand prints a command in help output.
func PrintCommand(name, args, desc string) {
	fmt.Fprintf(out, "    %s%s%s %s%s%s %s%s%s\n",
		Cyan, padRight(name, 12), Reset,
		Dim, padRight(args, 16), Reset,
		White, desc, Reset)
}

// PrintSection prints a section header.
func PrintSection(name string) {
	fmt.Fprintf(out, "  %s%s%s%s\n", Bold, BrightWhite, name, Reset)
}

// PrintBox prints content in a bordered box.
func PrintBox(title string, lines []string) {
	// Find the longest line
	maxLineLen := 0
	for _, line := range lines {
		if w := displayWidth(line); w > maxLineLen {
			maxLineLen = w
		}
	}

	// Inner width: " N. <line> " where N is single digit
	innerWidth := maxLineLen + 5

	// Make sure title fits
	if displayWidth(title)+4 > innerWidth {
		innerWidth = displayWidth(title) + 4
	}

	// Top border
	titlePadding := innerWidth - displayWidth(title) - 3
	fmt.Fprintf(out, "  %s╭─%s %s%s%s %s%s╮%s\n",
		Dim, Reset,
		Bold, title, Reset,
		Dim, strings.Repeat("─", titlePadding), Reset)

	// Content rows
	for i, line := range lines {
		linePadding := maxLineLen - displayWidth(line)
		fmt.Fprintf(out, "  %s│%s %s%d.%s %s%s %s│%s\n",
			Dim, Reset,
			Cyan, i+1, Reset,
			line, strings.Repeat(" ", linePadding),
			Dim, Reset)
	}

	// Bottom border
	fmt.Fprintf(out, "  %s╰%s╯%s\n", Dim, strings.Repeat("─", innerWidth), Reset)
}

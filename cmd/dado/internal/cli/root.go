package cli

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

// Run is the main entry point for the CLI.
func Run() {
	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "theme":
		RunTheme(os.Args[2:])
	case "component", "components":
		RunComponent(os.Args[2:])
	case "help", "-h", "--help":
		PrintUsage()
	case "version", "-v", "--version":
		lines := append(logoLines(), line(styled("  v"+cliVersion, cliMuted)), blank())
		renderCLILines(lines)
	default:
		renderCLILines([]cliLine{errorLine(fmt.Sprintf("Unknown command: %s", os.Args[1])), blank()})
		PrintUsage()
		os.Exit(1)
	}
}

// PrintUsage prints the main help message.
func PrintUsage() {
	lines := append(logoLines(),
		line(styled("  TUI application scaffolding tool", cliMuted)), blank(),
		sectionLine("USAGE"), line(styled("    dado", cliAccent), text(" <command> [arguments]")), blank(),
		sectionLine("COMMANDS"),
		commandLine("theme", "list|preview", "Manage themes"),
		commandLine("component", "list", "Browse available components"),
		commandLine("help", "", "Show this help message"),
		commandLine("version", "", "Show version"), blank(),
		sectionLine("EXAMPLES"), line(styled("    $", cliMuted), text(" dado theme preview nord")), blank(),
		sectionLine("LEARN MORE"),
		line(styled("    https://github.com/atterpac/dado", tcell.StyleDefault.Foreground(tcell.ColorBlue).Underline(true))), blank(),
	)
	renderCLILines(lines)
}

package cli

import (
	"fmt"
	"os"

	"github.com/atterpac/dado/cmd/dado/internal/ui"
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
		ui.PrintVersion()
	default:
		ui.PrintError(fmt.Sprintf("Unknown command: %s", os.Args[1]))
		fmt.Println()
		PrintUsage()
		os.Exit(1)
	}
}

// PrintUsage prints the main help message.
func PrintUsage() {
	ui.PrintLogo()
	fmt.Printf("  %s%sTUI application scaffolding tool%s\n\n", ui.Dim, ui.White, ui.Reset)

	fmt.Printf("  %s%sUSAGE%s\n", ui.Bold, ui.BrightWhite, ui.Reset)
	fmt.Printf("    %sdado%s <command> [arguments]\n\n", ui.Cyan, ui.Reset)

	fmt.Printf("  %s%sCOMMANDS%s\n", ui.Bold, ui.BrightWhite, ui.Reset)
	ui.PrintCommand("theme", "list|preview", "Manage themes")
	ui.PrintCommand("component", "list", "Browse available components")
	ui.PrintCommand("help", "", "Show this help message")
	ui.PrintCommand("version", "", "Show version")
	fmt.Println()

	fmt.Printf("  %s%sEXAMPLES%s\n", ui.Bold, ui.BrightWhite, ui.Reset)
	fmt.Printf("    %s$%s dado theme preview nord\n", ui.Dim, ui.Reset)
	fmt.Println()

	fmt.Printf("  %s%sLEARN MORE%s\n", ui.Bold, ui.BrightWhite, ui.Reset)
	fmt.Printf("    %shttps://github.com/atterpac/dado%s\n\n", ui.Underline+ui.Blue, ui.Reset)
}

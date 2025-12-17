package cli

import (
	"fmt"

	"github.com/atterpac/jig/cmd/jig/internal/scaffold"
	"github.com/atterpac/jig/cmd/jig/internal/ui"
)

// RunNew handles the "new" command.
func RunNew(args []string) {
	projectName := args[0]
	structured := false
	themeName := "tokyonight-night"
	interactive := true

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--simple":
			structured = false
			interactive = false
		case "--structured":
			structured = true
			interactive = false
		case "--theme":
			if i+1 < len(args) {
				themeName = args[i+1]
				i++
			}
		}
	}

	fmt.Println()
	ui.PrintLogo()
	fmt.Printf("  %sCreating new project: %s%s%s%s\n\n", ui.Dim, ui.Reset, ui.Bold, projectName, ui.Reset)

	// Interactive mode
	if interactive {
		structIdx := ui.SelectOption(
			"Project structure",
			[]string{"Simple", "Structured"},
			[]string{"Single main.go file", "cmd/internal layout"},
		)
		structured = structIdx == 1

		themes := []string{
			"tokyonight-night", "tokyonight-storm", "tokyonight-moon", "tokyonight-day",
			"catppuccin-mocha", "catppuccin-macchiato", "catppuccin-frappe", "catppuccin-latte",
			"dracula", "dracula-light",
			"gruvbox-dark", "gruvbox-light",
			"onedark", "onelight",
			"nord", "kanagawa", "monokai",
		}
		themeDescs := []string{
			"Dark blue (default)", "Storm variant", "Moon variant", "Light variant",
			"Dark pastel", "Medium dark", "Medium", "Light pastel",
			"Dark purple", "Light purple",
			"Retro dark", "Retro light",
			"Atom dark", "Atom light",
			"Arctic blue", "Japanese wave", "Classic dark",
		}
		themeIdx := ui.SelectOption("Default theme", themes, themeDescs)
		themeName = themes[themeIdx]
	}

	fmt.Println()
	ui.PrintInfo("Creating project structure...")
	fmt.Println()

	if structured {
		scaffold.CreateStructuredProject(projectName, themeName)
	} else {
		scaffold.CreateSimpleProject(projectName, themeName)
	}

	fmt.Println()
	ui.PrintSuccess(fmt.Sprintf("Created project %s%s%s", ui.Bold, projectName, ui.Reset))
	fmt.Println()

	// Print next steps
	ui.PrintBox("Next steps", []string{
		fmt.Sprintf("cd %s", projectName),
		"go mod tidy",
		"go run .",
	})
	fmt.Println()
}

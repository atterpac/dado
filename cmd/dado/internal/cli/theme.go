package cli

import (
	"fmt"
	"strings"

	ui "github.com/atterpac/dado/inline"
)

// ThemeInfo contains theme metadata for display.
type ThemeInfo struct {
	Name      string
	Desc      string
	Colors    []string
	IsDefault bool
}

// ThemeColors contains full color definitions for a theme.
type ThemeColors struct {
	Bg        string
	Fg        string
	FgDim     string
	Accent    string
	Success   string
	Warning   string
	Error     string
	Info      string
	Border    string
	Highlight string
}

var themeList = []ThemeInfo{
	{"dark", "Tokyo Night inspired", []string{"#1a1b26", "#c0caf5", "#7aa2f7", "#9ece6a"}, true},
	{"light", "Clean light mode", []string{"#ffffff", "#1a1b26", "#1e66f5", "#40a02b"}, false},
	{"catppuccin", "Pastel dark theme", []string{"#1e1e2e", "#cdd6f4", "#89b4fa", "#a6e3a1"}, false},
	{"nord", "Arctic inspired", []string{"#2e3440", "#eceff4", "#88c0d0", "#a3be8c"}, false},
	{"dracula", "Dark purple theme", []string{"#282a36", "#f8f8f2", "#bd93f9", "#50fa7b"}, false},
}

var themeColors = map[string]ThemeColors{
	"dark": {
		Bg: "#1a1b26", Fg: "#c0caf5", FgDim: "#565f89", Accent: "#7aa2f7",
		Success: "#9ece6a", Warning: "#e0af68", Error: "#f7768e", Info: "#7dcfff",
		Border: "#3b4261", Highlight: "#33467c",
	},
	"light": {
		Bg: "#ffffff", Fg: "#1a1b26", FgDim: "#6c7086", Accent: "#1e66f5",
		Success: "#40a02b", Warning: "#df8e1d", Error: "#d20f39", Info: "#04a5e5",
		Border: "#ccd0da", Highlight: "#e6e9ef",
	},
	"catppuccin": {
		Bg: "#1e1e2e", Fg: "#cdd6f4", FgDim: "#6c7086", Accent: "#89b4fa",
		Success: "#a6e3a1", Warning: "#f9e2af", Error: "#f38ba8", Info: "#89dceb",
		Border: "#45475a", Highlight: "#313244",
	},
	"nord": {
		Bg: "#2e3440", Fg: "#eceff4", FgDim: "#4c566a", Accent: "#88c0d0",
		Success: "#a3be8c", Warning: "#ebcb8b", Error: "#bf616a", Info: "#81a1c1",
		Border: "#4c566a", Highlight: "#3b4252",
	},
	"dracula": {
		Bg: "#282a36", Fg: "#f8f8f2", FgDim: "#6272a4", Accent: "#bd93f9",
		Success: "#50fa7b", Warning: "#f1fa8c", Error: "#ff5555", Info: "#8be9fd",
		Border: "#44475a", Highlight: "#44475a",
	},
}

// RunTheme handles the "theme" command.
func RunTheme(args []string) {
	if len(args) == 0 {
		printThemeUsage()
		return
	}

	switch args[0] {
	case "list":
		printThemeList()
	case "preview":
		if len(args) < 2 {
			ui.PrintError("Missing theme name")
			fmt.Printf("\n  %sUsage:%s dado theme preview <theme-name>\n\n", ui.Dim, ui.Reset)
			return
		}
		previewTheme(args[1])
	default:
		ui.PrintError(fmt.Sprintf("Unknown theme command: %s", args[0]))
	}
}

func printThemeUsage() {
	fmt.Println()
	ui.PrintLogo()
	fmt.Printf("  %sTheme management commands%s\n\n", ui.Dim, ui.Reset)
	fmt.Printf("  %s%sUSAGE%s\n", ui.Bold, ui.BrightWhite, ui.Reset)
	fmt.Printf("    %sdado theme%s <command>\n\n", ui.Cyan, ui.Reset)
	fmt.Printf("  %s%sCOMMANDS%s\n", ui.Bold, ui.BrightWhite, ui.Reset)
	ui.PrintCommand("list", "", "List all available themes")
	ui.PrintCommand("preview", "<name>", "Preview a theme's colors")
	fmt.Println()
}

func printThemeList() {
	fmt.Println()
	ui.PrintLogo()
	fmt.Printf("  %sAvailable themes%s\n\n", ui.Dim, ui.Reset)

	for _, t := range themeList {
		if t.IsDefault {
			fmt.Printf("  %s%s●%s %s%s%s %s(default)%s\n", ui.Bold, ui.Cyan, ui.Reset, ui.Bold, ui.Pad(t.Name, 12), ui.Reset, ui.Dim, ui.Reset)
		} else {
			fmt.Printf("  %s○%s %s%s%s\n", ui.Dim, ui.Reset, ui.White, ui.Pad(t.Name, 12), ui.Reset)
		}
		fmt.Printf("    %s%s%s\n", ui.Dim, t.Desc, ui.Reset)

		// Color swatches
		fmt.Print("    ")
		for _, c := range t.Colors {
			fmt.Printf("%s  %s", ui.ColorBg(c), ui.Reset)
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Printf("  %sTip:%s Run %sdado theme preview <name>%s for detailed view\n\n", ui.Dim, ui.Reset, ui.Cyan, ui.Reset)
}

func previewTheme(name string) {
	t, ok := themeColors[name]
	if !ok {
		ui.PrintError(fmt.Sprintf("Unknown theme: %s", name))
		fmt.Printf("\n  %sAvailable themes:%s dark, light, catppuccin, nord, dracula\n\n", ui.Dim, ui.Reset)
		return
	}

	fmt.Println()

	bg := ui.ColorBg(t.Bg)
	fg := ui.ColorFg(t.Fg)
	fgDim := ui.ColorFg(t.FgDim)
	accent := ui.ColorFg(t.Accent)
	border := ui.ColorFg(t.Border)

	width := 50

	// Top border
	fmt.Printf("  %s%s╭%s╮%s\n", bg, border, strings.Repeat("─", width-2), ui.Reset)

	// Title bar
	title := fmt.Sprintf(" %s Theme Preview ", strings.Title(name))
	titleW := ui.DisplayWidth(title)
	padding := (width - 2 - titleW) / 2
	fmt.Printf("  %s%s│%s%s%s%s%s%s│%s\n",
		bg, border,
		strings.Repeat(" ", padding), accent, title, fg, strings.Repeat(" ", width-2-titleW-padding),
		border, ui.Reset)

	// Separator
	fmt.Printf("  %s%s├%s┤%s\n", bg, border, strings.Repeat("─", width-2), ui.Reset)

	// Color entries
	colorOrder := []struct {
		key   string
		label string
		value string
	}{
		{"bg", "Background", t.Bg},
		{"fg", "Foreground", t.Fg},
		{"fg_dim", "Dimmed", t.FgDim},
		{"accent", "Accent", t.Accent},
		{"border", "Border", t.Border},
		{"highlight", "Highlight", t.Highlight},
		{"success", "Success", t.Success},
		{"warning", "Warning", t.Warning},
		{"error", "Error", t.Error},
		{"info", "Info", t.Info},
	}

	for _, c := range colorOrder {
		swatch := ui.ColorBg(c.value)
		labelColor := fg
		if c.key == "bg" || c.key == "fg_dim" || c.key == "border" || c.key == "highlight" {
			labelColor = fgDim
		}

		line := fmt.Sprintf(" %s%-12s%s %s  %s %s", labelColor, c.label, fg, swatch, ui.Reset+bg, c.value)
		padLen := width - 2 - 12 - 1 - 2 - 1 - 7 - 1
		fmt.Printf("  %s%s│%s%s│%s\n", bg, border, line, strings.Repeat(" ", padLen), ui.Reset)
	}

	// Bottom border
	fmt.Printf("  %s%s╰%s╯%s\n", bg, border, strings.Repeat("─", width-2), ui.Reset)
	fmt.Println()
}

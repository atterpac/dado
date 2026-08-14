package cli

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
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
	Bg, Fg, FgDim, Accent, Success, Warning, Error, Info, Border, Highlight string
}

var themeList = []ThemeInfo{
	{"dark", "Tokyo Night inspired", []string{"#1a1b26", "#c0caf5", "#7aa2f7", "#9ece6a"}, true},
	{"light", "Clean light mode", []string{"#ffffff", "#1a1b26", "#1e66f5", "#40a02b"}, false},
	{"catppuccin", "Pastel dark theme", []string{"#1e1e2e", "#cdd6f4", "#89b4fa", "#a6e3a1"}, false},
	{"nord", "Arctic inspired", []string{"#2e3440", "#eceff4", "#88c0d0", "#a3be8c"}, false},
	{"dracula", "Dark purple theme", []string{"#282a36", "#f8f8f2", "#bd93f9", "#50fa7b"}, false},
}

var themeColors = map[string]ThemeColors{
	"dark":       {"#1a1b26", "#c0caf5", "#565f89", "#7aa2f7", "#9ece6a", "#e0af68", "#f7768e", "#7dcfff", "#3b4261", "#33467c"},
	"light":      {"#ffffff", "#1a1b26", "#6c7086", "#1e66f5", "#40a02b", "#df8e1d", "#d20f39", "#04a5e5", "#ccd0da", "#e6e9ef"},
	"catppuccin": {"#1e1e2e", "#cdd6f4", "#6c7086", "#89b4fa", "#a6e3a1", "#f9e2af", "#f38ba8", "#89dceb", "#45475a", "#313244"},
	"nord":       {"#2e3440", "#eceff4", "#4c566a", "#88c0d0", "#a3be8c", "#ebcb8b", "#bf616a", "#81a1c1", "#4c566a", "#3b4252"},
	"dracula":    {"#282a36", "#f8f8f2", "#6272a4", "#bd93f9", "#50fa7b", "#f1fa8c", "#ff5555", "#8be9fd", "#44475a", "#44475a"},
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
			renderCLILines([]cliLine{errorLine("Missing theme name"), blank(), line(styled("  Usage: ", cliMuted), text("dado theme preview <theme-name>")), blank()})
			return
		}
		previewTheme(args[1])
	default:
		renderCLILines([]cliLine{errorLine(fmt.Sprintf("Unknown theme command: %s", args[0]))})
	}
}

func printThemeUsage() {
	lines := append([]cliLine{blank()}, logoLines()...)
	lines = append(lines,
		line(styled("  Theme management commands", cliMuted)), blank(), sectionLine("USAGE"),
		line(styled("    dado theme", cliAccent), text(" <command>")), blank(), sectionLine("COMMANDS"),
		commandLine("list", "", "List all available themes"),
		commandLine("preview", "<name>", "Preview a theme's colors"), blank(),
	)
	renderCLILines(lines)
}

func printThemeList() {
	lines := append([]cliLine{blank()}, logoLines()...)
	lines = append(lines, line(styled("  Available themes", cliMuted)), blank())
	for _, theme := range themeList {
		marker := "○"
		markerStyle := cliMuted
		suffix := ""
		if theme.IsDefault {
			marker, markerStyle, suffix = "●", cliAccent.Bold(true), " (default)"
		}
		lines = append(lines,
			line(styled("  "+marker+" ", markerStyle), styled(padASCII(theme.Name, 12), cliText.Bold(theme.IsDefault)), styled(suffix, cliMuted)),
			line(styled("    "+theme.Desc, cliMuted)),
		)
		swatches := cliLine{styled("    ", cliText)}
		for _, color := range theme.Colors {
			swatches = append(swatches, styled("  ", tcell.StyleDefault.Background(terminalColor(color))), text(" "))
		}
		lines = append(lines, swatches, blank())
	}
	lines = append(lines, line(styled("  Tip: ", cliMuted), text("Run "), styled("dado theme preview <name>", cliAccent), text(" for detailed view")), blank())
	renderCLILines(lines)
}

func previewTheme(name string) {
	theme, ok := themeColors[name]
	if !ok {
		renderCLILines([]cliLine{errorLine(fmt.Sprintf("Unknown theme: %s", name)), blank(), line(styled("  Available themes: ", cliMuted), text("dark, light, catppuccin, nord, dracula")), blank()})
		return
	}

	const width = 50
	background := terminalColor(theme.Bg)
	foreground := terminalColor(theme.Fg)
	dimmed := terminalColor(theme.FgDim)
	accent := terminalColor(theme.Accent)
	border := terminalColor(theme.Border)
	bgStyle := tcell.StyleDefault.Foreground(foreground).Background(background)
	borderStyle := bgStyle.Foreground(border)

	titleName := name
	if titleName != "" {
		titleName = strings.ToUpper(titleName[:1]) + titleName[1:]
	}
	title := " " + titleName + " Theme Preview "
	left := (width - 2 - len(title)) / 2
	right := width - 2 - len(title) - left
	lines := []cliLine{blank(),
		line(text("  "), styled("╭"+strings.Repeat("─", width-2)+"╮", borderStyle)),
		line(text("  "), styled("│"+strings.Repeat(" ", left), borderStyle), styled(title, bgStyle.Foreground(accent).Bold(true)), styled(strings.Repeat(" ", right)+"│", borderStyle)),
		line(text("  "), styled("├"+strings.Repeat("─", width-2)+"┤", borderStyle)),
	}
	colors := []struct{ label, value string }{
		{"Background", theme.Bg}, {"Foreground", theme.Fg}, {"Dimmed", theme.FgDim},
		{"Accent", theme.Accent}, {"Border", theme.Border}, {"Highlight", theme.Highlight},
		{"Success", theme.Success}, {"Warning", theme.Warning}, {"Error", theme.Error}, {"Info", theme.Info},
	}
	for _, color := range colors {
		labelStyle := bgStyle
		if color.label == "Background" || color.label == "Dimmed" || color.label == "Border" || color.label == "Highlight" {
			labelStyle = bgStyle.Foreground(dimmed)
		}
		contentWidth := 1 + 12 + 1 + 2 + 1 + len(color.value)
		lines = append(lines, line(
			text("  "), styled("│", borderStyle), styled(" "+padASCII(color.label, 12)+" ", labelStyle),
			styled("  ", tcell.StyleDefault.Background(terminalColor(color.value))), styled(" "+color.value, bgStyle),
			styled(strings.Repeat(" ", width-2-contentWidth)+"│", borderStyle),
		))
	}
	lines = append(lines, line(text("  "), styled("╰"+strings.Repeat("─", width-2)+"╯", borderStyle)), blank())
	renderCLILines(lines)
}

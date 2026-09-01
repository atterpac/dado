package inline

import (
	"regexp"
	"strings"

	"github.com/rivo/uniseg"
)

// ansiRE matches SGR sequences and OSC 8 hyperlinks. Renderer persistent
// output strips these when writing to a non-terminal destination.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m|\x1b\]8;;[^\x1b]*\x1b\\`)

func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

// displayWidth returns the number of terminal cells occupied by s.
func displayWidth(s string) int { return uniseg.StringWidth(stripANSI(s)) }

func splitGraphemes(text string) []string {
	graphemes := uniseg.NewGraphemes(text)
	result := make([]string, 0, uniseg.GraphemeClusterCount(text))
	for graphemes.Next() {
		result = append(result, graphemes.Str())
	}
	return result
}

func graphemesWidth(graphemes []string) int {
	width := 0
	for _, grapheme := range graphemes {
		width += uniseg.StringWidth(grapheme)
	}
	return width
}

func truncateCells(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if displayWidth(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}

	var result strings.Builder
	used := 0
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		clusterWidth := graphemes.Width()
		if used+clusterWidth > width-1 {
			break
		}
		result.WriteString(graphemes.Str())
		used += clusterWidth
	}
	result.WriteString("…")
	return result.String()
}

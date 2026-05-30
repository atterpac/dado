package util

import (
	"fmt"
	"strings"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/dado/theme"
)

// ApplyHorizontalGradient applies a left-to-right gradient to text.
// Colors should be hex strings like "#ff0000".
func ApplyHorizontalGradient(text string, colors []string) string {
	if len(colors) == 0 {
		return text
	}
	if len(colors) == 1 {
		return "[" + colors[0] + "]" + text + "[-]"
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		runes := []rune(line)
		if len(runes) == 0 {
			result = append(result, "")
			continue
		}

		var coloredLine strings.Builder
		for i, r := range runes {
			position := float64(i) / float64(len(runes)-1)
			if len(runes) == 1 {
				position = 0
			}
			color := InterpolateColors(colors, position)
			coloredLine.WriteString("[" + color + "]")
			coloredLine.WriteRune(r)
			coloredLine.WriteString("[-]")
		}
		result = append(result, coloredLine.String())
	}

	return strings.Join(result, "\n")
}

// ApplyVerticalGradient applies a top-to-bottom gradient to text.
// Text is split by newlines, each line gets a color based on position.
func ApplyVerticalGradient(text string, colors []string) string {
	if len(colors) == 0 {
		return text
	}
	if len(colors) == 1 {
		return "[" + colors[0] + "]" + text + "[-]"
	}

	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	var result []string
	for i, line := range lines {
		position := float64(i) / float64(len(lines)-1)
		if len(lines) == 1 {
			position = 0
		}
		color := InterpolateColors(colors, position)
		result = append(result, "["+color+"]"+line+"[-]")
	}

	return strings.Join(result, "\n")
}

// ApplyDiagonalGradient applies a diagonal (top-left to bottom-right) gradient.
func ApplyDiagonalGradient(text string, colors []string) string {
	if len(colors) == 0 {
		return text
	}
	if len(colors) == 1 {
		return "[" + colors[0] + "]" + text + "[-]"
	}

	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	// Find max line length for normalization
	maxLen := 0
	for _, line := range lines {
		if len([]rune(line)) > maxLen {
			maxLen = len([]rune(line))
		}
	}

	var result []string
	for row, line := range lines {
		runes := []rune(line)
		if len(runes) == 0 {
			result = append(result, "")
			continue
		}

		var coloredLine strings.Builder
		for col, r := range runes {
			// Diagonal position: average of row and column positions
			rowPos := float64(row) / float64(max(len(lines)-1, 1))
			colPos := float64(col) / float64(max(maxLen-1, 1))
			position := (rowPos + colPos) / 2

			color := InterpolateColors(colors, position)
			coloredLine.WriteString("[" + color + "]")
			coloredLine.WriteRune(r)
			coloredLine.WriteString("[-]")
		}
		result = append(result, coloredLine.String())
	}

	return strings.Join(result, "\n")
}

// ApplyReverseDiagonalGradient applies a reverse diagonal gradient.
func ApplyReverseDiagonalGradient(text string, colors []string) string {
	if len(colors) == 0 {
		return text
	}
	if len(colors) == 1 {
		return "[" + colors[0] + "]" + text + "[-]"
	}

	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	// Find max line length for normalization
	maxLen := 0
	for _, line := range lines {
		if len([]rune(line)) > maxLen {
			maxLen = len([]rune(line))
		}
	}

	var result []string
	for row, line := range lines {
		runes := []rune(line)
		if len(runes) == 0 {
			result = append(result, "")
			continue
		}

		var coloredLine strings.Builder
		for col, r := range runes {
			// Reverse diagonal: row position + (1 - column position)
			rowPos := float64(row) / float64(max(len(lines)-1, 1))
			colPos := 1 - float64(col)/float64(max(maxLen-1, 1))
			position := (rowPos + colPos) / 2

			color := InterpolateColors(colors, position)
			coloredLine.WriteString("[" + color + "]")
			coloredLine.WriteRune(r)
			coloredLine.WriteString("[-]")
		}
		result = append(result, coloredLine.String())
	}

	return strings.Join(result, "\n")
}

// InterpolateColor returns a color between c1 and c2 based on t (0.0 to 1.0).
// Colors should be hex strings like "#ff0000".
func InterpolateColor(c1, c2 string, t float64) string {
	r1, g1, b1 := parseHexColor(c1)
	r2, g2, b2 := parseHexColor(c2)

	r := int(float64(r1) + t*(float64(r2)-float64(r1)))
	g := int(float64(g1) + t*(float64(g2)-float64(g1)))
	b := int(float64(b1) + t*(float64(b2)-float64(b1)))

	return fmt.Sprintf("#%02x%02x%02x", clamp(r, 0, 255), clamp(g, 0, 255), clamp(b, 0, 255))
}

// InterpolateColors returns a color from a gradient based on position.
// Position is 0.0 to 1.0, colors array defines gradient stops.
func InterpolateColors(colors []string, position float64) string {
	if len(colors) == 0 {
		return "#ffffff"
	}
	if len(colors) == 1 {
		return colors[0]
	}

	// Clamp position
	if position <= 0 {
		return colors[0]
	}
	if position >= 1 {
		return colors[len(colors)-1]
	}

	// Find which segment we're in
	segments := len(colors) - 1
	scaledPos := position * float64(segments)
	segment := int(scaledPos)
	if segment >= segments {
		segment = segments - 1
	}
	segmentPos := scaledPos - float64(segment)

	return InterpolateColor(colors[segment], colors[segment+1], segmentPos)
}

// DefaultGradientColors returns theme-based gradient colors.
func DefaultGradientColors() []string {
	return []string{
		theme.TagAccent(),
		theme.TagSuccess(),
		theme.TagFgDim(),
	}
}

// AccentGradientColors returns accent-based gradient colors.
func AccentGradientColors() []string {
	return []string{
		theme.TagAccent(),
		theme.TagInfo(),
		theme.TagSuccess(),
	}
}

// parseHexColor parses a hex color string like "#ff0000" to RGB components.
func parseHexColor(hex string) (r, g, b int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 255, 255, 255
	}

	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// clamp restricts a value to the given range.
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

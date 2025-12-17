package theme

import (
	"fmt"
	"strings"
)

// GradientType defines the direction of gradient application.
type GradientType int

const (
	GradientDiagonal GradientType = iota // Top-left to bottom-right
	GradientHorizontal                   // Left to right
	GradientVertical                     // Top to bottom
	GradientReverseDiagonal              // Top-right to bottom-left
)

// String returns a human-readable name for the gradient type.
func (g GradientType) String() string {
	switch g {
	case GradientDiagonal:
		return "Diagonal"
	case GradientHorizontal:
		return "Horizontal"
	case GradientVertical:
		return "Vertical"
	case GradientReverseDiagonal:
		return "Reverse Diagonal"
	default:
		return "Unknown"
	}
}

// Next returns the next gradient type in sequence, wrapping around.
func (g GradientType) Next() GradientType {
	return (g + 1) % 4
}

// DefaultGradientColors returns theme-based gradient colors (Accent -> Success -> FgDim).
func DefaultGradientColors() []string {
	return []string{
		TagAccent(),
		TagSuccess(),
		TagFgDim(),
	}
}

// AccentGradientColors returns accent-focused gradient colors.
func AccentGradientColors() []string {
	return []string{
		TagAccent(),
		TagHighlight(),
		TagSuccess(),
	}
}

// hexToRGB parses a hex color string like "#f5c2e7" into RGB components.
func hexToRGB(hex string) (r, g, b int) {
	if len(hex) == 7 && hex[0] == '#' {
		fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	}
	return
}

// rgbToHexString converts RGB values to a hex string.
func rgbToHexString(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// InterpolateHex blends two hex colors based on a ratio (0.0 to 1.0).
func InterpolateHex(hex1, hex2 string, ratio float64) string {
	r1, g1, b1 := hexToRGB(hex1)
	r2, g2, b2 := hexToRGB(hex2)

	r := int(float64(r1) + ratio*(float64(r2)-float64(r1)))
	g := int(float64(g1) + ratio*(float64(g2)-float64(g1)))
	b := int(float64(b1) + ratio*(float64(b2)-float64(b1)))

	return rgbToHexString(r, g, b)
}

// InterpolateGradient returns a color at position t (0.0 to 1.0) across an N-color gradient.
func InterpolateGradient(colors []string, t float64) string {
	if len(colors) == 0 {
		return "#ffffff"
	}
	if len(colors) == 1 {
		return colors[0]
	}

	// Clamp t to [0, 1]
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// For N colors, we have N-1 segments
	segments := len(colors) - 1
	segmentSize := 1.0 / float64(segments)

	// Find which segment t falls into
	segment := int(t / segmentSize)
	if segment >= segments {
		segment = segments - 1
	}

	// Calculate position within the segment
	segmentStart := float64(segment) * segmentSize
	localT := (t - segmentStart) / segmentSize

	return InterpolateHex(colors[segment], colors[segment+1], localT)
}

// ApplyGradient applies the specified gradient type to ASCII art text.
func ApplyGradient(text string, gradientType GradientType, colors []string) string {
	switch gradientType {
	case GradientDiagonal:
		return ApplyDiagonalGradient(text, colors)
	case GradientHorizontal:
		return ApplyHorizontalGradient(text, colors)
	case GradientVertical:
		return ApplyVerticalGradient(text, colors)
	case GradientReverseDiagonal:
		return ApplyReverseDiagonalGradient(text, colors)
	default:
		return ApplyDiagonalGradient(text, colors)
	}
}

// ApplyHorizontalGradient applies a left-to-right gradient to ASCII art.
func ApplyHorizontalGradient(text string, colors []string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || len(colors) == 0 {
		return text
	}

	// Find max line width
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	if maxWidth == 0 {
		return text
	}

	segments := 10
	segmentWidth := maxWidth / segments
	if segmentWidth < 1 {
		segmentWidth = 1
	}

	var result strings.Builder
	for i, line := range lines {
		runes := []rune(line)
		currentSegment := -1
		for j, r := range runes {
			segment := j / segmentWidth
			if segment >= segments {
				segment = segments - 1
			}

			if segment != currentSegment {
				if currentSegment >= 0 {
					result.WriteString("[-]")
				}
				t := float64(segment) / float64(segments-1)
				color := InterpolateGradient(colors, t)
				result.WriteString(fmt.Sprintf("[%s]", color))
				currentSegment = segment
			}
			result.WriteRune(r)
		}
		if currentSegment >= 0 {
			result.WriteString("[-]")
		}
		if i < len(lines)-1 {
			result.WriteRune('\n')
		}
	}

	return result.String()
}

// ApplyVerticalGradient applies a top-to-bottom gradient to ASCII art.
func ApplyVerticalGradient(text string, colors []string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || len(colors) == 0 {
		return text
	}

	var result strings.Builder
	for i, line := range lines {
		t := float64(i) / float64(len(lines)-1)
		if len(lines) == 1 {
			t = 0.5
		}
		color := InterpolateGradient(colors, t)
		result.WriteString(fmt.Sprintf("[%s]%s[-]", color, line))
		if i < len(lines)-1 {
			result.WriteRune('\n')
		}
	}

	return result.String()
}

// ApplyDiagonalGradient applies a diagonal gradient (top-left to bottom-right) to ASCII art.
func ApplyDiagonalGradient(text string, colors []string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || len(colors) == 0 {
		return text
	}

	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	if maxWidth == 0 {
		return text
	}

	bands := 12
	maxDiag := maxWidth + len(lines)
	bandSize := maxDiag / bands
	if bandSize < 1 {
		bandSize = 1
	}

	var result strings.Builder
	for i, line := range lines {
		runes := []rune(line)
		currentBand := -1
		for j, r := range runes {
			diagPos := i + j
			band := diagPos / bandSize
			if band >= bands {
				band = bands - 1
			}

			if band != currentBand {
				if currentBand >= 0 {
					result.WriteString("[-]")
				}
				t := float64(band) / float64(bands-1)
				color := InterpolateGradient(colors, t)
				result.WriteString(fmt.Sprintf("[%s]", color))
				currentBand = band
			}
			result.WriteRune(r)
		}
		if currentBand >= 0 {
			result.WriteString("[-]")
		}
		if i < len(lines)-1 {
			result.WriteRune('\n')
		}
	}

	return result.String()
}

// ApplyReverseDiagonalGradient applies a diagonal gradient (top-right to bottom-left) to ASCII art.
func ApplyReverseDiagonalGradient(text string, colors []string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || len(colors) == 0 {
		return text
	}

	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	if maxWidth == 0 {
		return text
	}

	bands := 12
	maxDiag := maxWidth + len(lines)
	bandSize := maxDiag / bands
	if bandSize < 1 {
		bandSize = 1
	}

	var result strings.Builder
	for i, line := range lines {
		runes := []rune(line)
		currentBand := -1
		for j, r := range runes {
			diagPos := i + (maxWidth - j)
			band := diagPos / bandSize
			if band >= bands {
				band = bands - 1
			}

			if band != currentBand {
				if currentBand >= 0 {
					result.WriteString("[-]")
				}
				t := float64(band) / float64(bands-1)
				color := InterpolateGradient(colors, t)
				result.WriteString(fmt.Sprintf("[%s]", color))
				currentBand = band
			}
			result.WriteRune(r)
		}
		if currentBand >= 0 {
			result.WriteString("[-]")
		}
		if i < len(lines)-1 {
			result.WriteRune('\n')
		}
	}

	return result.String()
}

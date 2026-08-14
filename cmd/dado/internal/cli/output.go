package cli

import (
	"os"
	"strings"

	"github.com/atterpac/dado/inline"
	"github.com/gdamore/tcell/v2"
)

const cliVersion = "0.1.0"

var (
	cliText    = tcell.StyleDefault
	cliMuted   = tcell.StyleDefault.Foreground(tcell.ColorGray)
	cliAccent  = tcell.StyleDefault.Foreground(tcell.ColorAqua)
	cliHeading = tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true)
	cliError   = tcell.StyleDefault.Foreground(tcell.ColorRed)
	cliLogo    = tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true)
)

type cliSpan struct {
	text  string
	style tcell.Style
}

type cliLine []cliSpan

func line(spans ...cliSpan) cliLine { return spans }
func text(value string) cliSpan     { return cliSpan{text: value, style: cliText} }
func styled(value string, style tcell.Style) cliSpan {
	return cliSpan{text: value, style: style}
}
func blank() cliLine { return nil }

func logoLines() []cliLine {
	return []cliLine{
		line(styled("    ┌──╮  ╭──╮  ┌──╮  ╭──╮", cliLogo)),
		line(styled("    │  │  ├──┤  │  │  │  │", cliLogo)),
		line(styled("    └──╯  ╵  ╵  └──╯  ╰──╯", cliLogo)),
	}
}

func errorLine(message string) cliLine {
	return line(styled("  ✗ "+message, cliError.Bold(true)))
}

func sectionLine(name string) cliLine {
	return line(styled("  "+name, cliHeading))
}

func commandLine(name, args, description string) cliLine {
	return line(
		styled("    "+padASCII(name, 12)+" ", cliAccent),
		styled(padASCII(args, 16)+" ", cliMuted),
		text(description),
	)
}

func renderCLILines(lines []cliLine) {
	frame := inline.NewFrame(80, len(lines))
	for y, spans := range lines {
		x := 0
		for _, span := range spans {
			remaining := span.text
			for remaining != "" {
				rest, width := frame.Put(x, y, remaining, span.style)
				if width <= 0 || rest == remaining {
					break
				}
				x += width
				remaining = rest
			}
		}
	}
	renderer := inline.NewRenderer(inline.WithOutput(os.Stdout))
	_ = renderer.Render(frame)
	_ = renderer.Close()
}

func padASCII(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}

func terminalColor(hex string) tcell.Color { return tcell.GetColor(hex) }

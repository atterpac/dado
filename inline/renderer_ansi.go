package inline

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func writeStyledRow(buf *bytes.Buffer, row []frameCell) {
	last := lastVisibleCell(row)
	var active tcell.Style
	haveStyle := false
	for x := 0; x < last; {
		cell := row[x]
		if !haveStyle || cell.style != active {
			buf.WriteString(ansiReset)
			buf.WriteString(styleSequence(cell.style))
			active = cell.style
			haveStyle = true
		}
		width := cell.width
		if width < 1 {
			width = 1
		}
		if x+width > len(row) {
			buf.WriteByte(' ')
		} else {
			buf.WriteString(cell.text)
		}
		x += width
	}
	if haveStyle {
		buf.WriteString(ansiReset)
	}
}

func lastVisibleCell(row []frameCell) int {
	last := 0
	for x := 0; x < len(row); {
		cell := row[x]
		width := cell.width
		if width < 1 {
			width = 1
		}
		if cell.text != " " || cell.style != tcell.StyleDefault {
			last = min(len(row), x+width)
		}
		x += width
	}
	return last
}

func styleSequence(style tcell.Style) string {
	fg, bg, attrs := style.Decompose()
	codes := make([]string, 0, 10)
	if attrs&tcell.AttrBold != 0 {
		codes = append(codes, "1")
	}
	if attrs&tcell.AttrDim != 0 {
		codes = append(codes, "2")
	}
	if attrs&tcell.AttrItalic != 0 {
		codes = append(codes, "3")
	}
	if attrs&tcell.AttrUnderline != 0 {
		underline := style.GetUnderlineStyle()
		if underline <= tcell.UnderlineStyleSolid {
			codes = append(codes, "4")
		} else {
			codes = append(codes, "4:"+strconv.Itoa(int(underline)))
		}
	}
	if attrs&tcell.AttrBlink != 0 {
		codes = append(codes, "5")
	}
	if attrs&tcell.AttrReverse != 0 {
		codes = append(codes, "7")
	}
	if attrs&tcell.AttrStrikeThrough != 0 {
		codes = append(codes, "9")
	}
	if code := colorCode(fg, false); code != "" {
		codes = append(codes, code)
	}
	if code := colorCode(bg, true); code != "" {
		codes = append(codes, code)
	}
	if color := style.GetUnderlineColor(); color.Valid() {
		r, g, b := color.RGB()
		if r >= 0 {
			codes = append(codes, fmt.Sprintf("58;2;%d;%d;%d", r, g, b))
		}
	}
	if len(codes) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

func colorCode(color tcell.Color, background bool) string {
	if !color.Valid() {
		return ""
	}
	if color.IsRGB() {
		r, g, b := color.RGB()
		if r < 0 {
			return ""
		}
		prefix := 38
		if background {
			prefix = 48
		}
		return fmt.Sprintf("%d;2;%d;%d;%d", prefix, r, g, b)
	}

	index := int(uint64(color &^ tcell.ColorValid))
	switch {
	case index < 8:
		base := 30
		if background {
			base = 40
		}
		return strconv.Itoa(base + index)
	case index < 16:
		base := 90
		if background {
			base = 100
		}
		return strconv.Itoa(base + index - 8)
	case index < 256:
		prefix := 38
		if background {
			prefix = 48
		}
		return fmt.Sprintf("%d;5;%d", prefix, index)
	default:
		return ""
	}
}

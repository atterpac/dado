package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// Label is a simple text display component.
// It wraps core.TextView with themed defaults and a cleaner API.
type Label struct {
	*core.TextView
	subs Subscriptions
}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (l *Label) Subs() *Subscriptions { return &l.subs }

// NewLabel creates a new Label with the given text.
func NewLabel(text string) *Label {
	tv := core.NewTextView()
	tv.SetText(text)
	tv.SetDynamicColors(true)
	tv.SetBackgroundColor(theme.Bg())
	tv.SetTextColor(theme.Fg())

	l := &Label{
		TextView: tv,
	}

	l.subs.Add(theme.RegisterFn(func(c tcell.Color) { tv.SetBackgroundColor(c) }))

	return l
}

// SetText sets the label text.
func (l *Label) SetText(text string) *Label {
	l.TextView.SetText(text)
	return l
}

// SetAlign sets the text alignment.
func (l *Label) SetAlign(align Align) *Label {
	var coreAlign int
	switch align {
	case AlignLeft:
		coreAlign = core.AlignLeft
	case AlignRight:
		coreAlign = core.AlignRight
	default:
		coreAlign = core.AlignCenter
	}
	l.TextView.SetTextAlign(coreAlign)
	return l
}

// SetColor sets the text color.
func (l *Label) SetColor(color tcell.Color) *Label {
	l.TextView.SetTextColor(color)
	return l
}

// SetBold sets whether the text is bold.
func (l *Label) SetBold(bold bool) *Label {
	if bold {
		l.TextView.SetTextStyle(tcell.StyleDefault.Bold(true))
	} else {
		l.TextView.SetTextStyle(tcell.StyleDefault)
	}
	return l
}

// SetWordWrap enables or disables word wrapping.
func (l *Label) SetWordWrap(wrap bool) *Label {
	l.TextView.SetWordWrap(wrap)
	return l
}

// SetScrollable enables or disables scrolling.
func (l *Label) SetScrollable(scrollable bool) *Label {
	l.TextView.SetScrollable(scrollable)
	return l
}

// SetDynamicColors enables or disables dynamic color tags.
// When enabled, you can use [color]text[-] syntax.
func (l *Label) SetDynamicColors(enabled bool) *Label {
	l.TextView.SetDynamicColors(enabled)
	return l
}

// SetRegions enables or disables region tags.
// When enabled, you can use ["region"]text[""] syntax for clickable regions.
func (l *Label) SetRegions(enabled bool) *Label {
	l.TextView.SetRegions(enabled)
	return l
}

// Text is a convenience function to create a Label.
// Alias for NewLabel.
func Text(text string) *Label {
	return NewLabel(text)
}

// Focus implements core.Widget. Focuses the underlying core.TextView.
func (l *Label) Focus() {
	l.TextView.Focus()
}

// Blur implements core.Widget.
func (l *Label) Blur() {
	l.TextView.Blur()
}

// Rect implements core.Widget.
func (l *Label) Rect() (x, y, w, h int) {
	x, y, w, h = l.TextView.GetRect()
	return
}

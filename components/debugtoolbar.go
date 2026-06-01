package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// DebugToolKind classifies how a DebugTool occupies the screen and consumes input.
type DebugToolKind int

const (
	// DebugToolPanel docks a navigable panel and captures input while active —
	// nav keys drive the panel, the app underneath is not interactive.
	DebugToolPanel DebugToolKind = iota
	// DebugToolInline draws a small HUD and passes input through to the app, so
	// the real UI stays interactive while the tool observes it (e.g. cell probe).
	DebugToolInline
)

// DebugTool is one entry on the debug toolbar. Exactly one tool is active while
// the toolbar is visible; Tab cycles between them.
type DebugTool interface {
	// Name is the short label shown on the toolbar strip.
	Name() string
	// Kind reports whether the tool docks a capturing panel or an inline HUD.
	Kind() DebugToolKind
	// Activate is called when the tool becomes the active one (set up bus
	// subscriptions, enable observers, etc.).
	Activate()
	// Deactivate is called when the tool stops being active (tear down).
	Deactivate()
	// Draw renders the tool into the given region. For panel tools the region is
	// the docked panel rect; for inline tools it is the area available above the
	// strip, and the tool self-positions a small HUD within it.
	Draw(screen tcell.Screen, x, y, w, h int)
	// HandleKey processes a key while the tool is active. Panel tools return true
	// to consume; the return is ignored for inline tools (always passthrough).
	HandleKey(ev *tcell.EventKey) bool
	// Hints are shown on the menu/hint bar while the tool is active.
	Hints() []KeyHint
}

// DockSide selects where the strip sits and where panels open. Only DockBottom
// (strip bottom, panel right) is implemented in v1; the field exists so re-dock
// can be added without reshaping the controller.
type DockSide int

const (
	DockBottom DockSide = iota
)

// DebugToolbar is a floating debug surface drawn on top of the live app via
// App.SetAfterDrawFunc, driven by keys routed from the app's input capture. It
// does not replace the app view (unlike a full-screen page), so you can see and
// — with inline tools — interact with the UI being debugged.
//
// While visible, exactly one tool is active. Tab/BackTab cycle tools, Esc hides
// the toolbar. Panel tools capture input; inline tools let it pass through.
type DebugToolbar struct {
	app     *core.App
	tools   []DebugTool
	active  int
	visible bool
	dock    DockSide
}

// NewDebugToolbar builds the toolbar with the default tool set: event log
// (panel), widget-tree inspector (panel), and cell probe (inline).
func NewDebugToolbar(app *core.App) *DebugToolbar {
	t := &DebugToolbar{app: app, dock: DockBottom}
	t.tools = []DebugTool{
		NewEventLogTool(0),
		NewWidgetTreeTool(app),
		NewCellProbeTool(app),
	}
	return t
}

// Visible reports whether the toolbar is currently shown.
func (t *DebugToolbar) Visible() bool { return t.visible }

// Toggle shows or hides the toolbar, activating/deactivating the active tool.
func (t *DebugToolbar) Toggle() {
	if t.visible {
		t.hide()
	} else {
		t.show()
	}
}

func (t *DebugToolbar) show() {
	if t.visible || len(t.tools) == 0 {
		return
	}
	t.visible = true
	t.tools[t.active].Activate()
}

func (t *DebugToolbar) hide() {
	if !t.visible {
		return
	}
	t.tools[t.active].Deactivate()
	t.visible = false
}

// ActiveHints returns the active tool's hints plus the toolbar-global ones.
func (t *DebugToolbar) ActiveHints() []KeyHint {
	hints := []KeyHint{
		{Key: "Tab", Description: "Next tool"},
		{Key: "Esc", Description: "Close debug"},
	}
	if t.visible && len(t.tools) > 0 {
		hints = append(hints, t.tools[t.active].Hints()...)
	}
	return hints
}

func (t *DebugToolbar) cycle(delta int) {
	if len(t.tools) == 0 {
		return
	}
	t.tools[t.active].Deactivate()
	t.active = (t.active + delta + len(t.tools)) % len(t.tools)
	t.tools[t.active].Activate()
}

// HandleKey processes a key while the toolbar is visible. The bool reports
// whether the event was consumed (true) or should pass through to the app
// (false). Toolbar-global keys (Tab/BackTab/Esc) and panel tools consume;
// inline tools observe but pass through so the app stays interactive.
func (t *DebugToolbar) HandleKey(ev *tcell.EventKey) bool {
	if !t.visible || len(t.tools) == 0 {
		return false
	}
	switch ev.Key() {
	case tcell.KeyTab:
		t.cycle(+1)
		return true
	case tcell.KeyBacktab:
		t.cycle(-1)
		return true
	case tcell.KeyEscape:
		t.hide()
		return true
	}
	tool := t.tools[t.active]
	tool.HandleKey(ev)
	if tool.Kind() == DebugToolInline {
		return false // passthrough — app still receives the key
	}
	return true // panel tools are modal: swallow everything else
}

// Draw paints the toolbar on top of the app. Intended for App.SetAfterDrawFunc.
func (t *DebugToolbar) Draw(screen tcell.Screen) {
	if !t.visible || len(t.tools) == 0 {
		return
	}
	w, h := screen.Size()
	if w <= 0 || h <= 0 {
		return
	}
	stripY := h - 1
	tool := t.tools[t.active]

	// Tool content above the strip.
	if tool.Kind() == DebugToolPanel {
		pw := w * 2 / 5
		if pw < 32 {
			pw = 32
		}
		if pw > w {
			pw = w
		}
		tool.Draw(screen, w-pw, 0, pw, stripY)
	} else {
		tool.Draw(screen, 0, 0, w, stripY)
	}

	t.drawStrip(screen, stripY, w)
}

// drawStrip renders the single-row tool selector at the bottom.
func (t *DebugToolbar) drawStrip(screen tcell.Screen, y, w int) {
	barStyle := tcell.StyleDefault.Background(theme.BgDark()).Foreground(theme.FgDim())
	activeStyle := tcell.StyleDefault.Background(theme.Accent()).Foreground(theme.Bg()).Bold(true)
	sepStyle := tcell.StyleDefault.Background(theme.BgDark()).Foreground(theme.FgMuted())

	core.FillRect(screen, 0, y, w, 1, ' ', barStyle)

	strip := core.NewText()
	strip.Append(" ", barStyle)
	for i, tool := range t.tools {
		if i > 0 {
			strip.Append(" │ ", sepStyle)
		}
		label := " " + tool.Name() + " "
		if i == t.active {
			strip.Append(label, activeStyle)
		} else {
			strip.Append(tool.Name(), barStyle)
		}
	}

	mode := "passthrough"
	if t.tools[t.active].Kind() == DebugToolPanel {
		mode = "capture"
	}
	hint := "  " + mode + " · tab:cycle esc:close"

	strip.Draw(screen, 0, y, w)
	// Right-align the mode/hint if room.
	hx := w - len(hint)
	if hx > strip.Width()+1 {
		core.PrintAt(screen, hx, y, hint, sepStyle)
	}
}

package components

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/bus"
	"github.com/atterpac/jig/theme"
)

// DebugOverlay is a runtime view of bus events. Subscribes to bus.Default()
// while active and renders the most recent events in a table. Supports
// source-letter filters (b/t/n/a/i/e for binding/theme/nav/async/input/effect)
// toggled by the corresponding key. 'c' clears the filter, 'p' pauses live
// updates, 'x' clears the table.
//
// The overlay enables the bus on Start and restores its prior state on Stop,
// so callers do not need to manage bus.SetEnabled themselves.
type DebugOverlay struct {
	*tview.Table
	base *ComponentBase

	mu          sync.Mutex
	events      []bus.Event // last N, newest last
	cap         int
	sources     map[string]bool // empty = all
	paused      bool
	prevEnabled bool
	onClose     func()
}

// NewDebugOverlay constructs a debug overlay backed by the default bus.
// capacity bounds the visible event list (defaults to 500 if <= 0).
func NewDebugOverlay(capacity int) *DebugOverlay {
	if capacity <= 0 {
		capacity = 500
	}
	tbl := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)
	tbl.SetBorder(true).SetTitle(" Bus Debug ")

	d := &DebugOverlay{
		Table:   tbl,
		cap:     capacity,
		sources: map[string]bool{},
	}

	d.base = NewComponentBase(tbl).
		SetName("debug-overlay").
		SetHints([]KeyHint{
			{Key: "b/t/n/a/i/e", Description: "Toggle source filter"},
			{Key: "c", Description: "Clear filter"},
			{Key: "p", Description: "Pause/resume"},
			{Key: "x", Description: "Clear events"},
			{Key: "Esc", Description: "Close"},
		}).
		SetOnStart(d.onStart).
		SetOnStop(d.onStop).
		SetInputHandler(d.handleKey)

	d.renderHeader()
	return d
}

// Base returns the component base for nav lifecycle integration.
func (d *DebugOverlay) Base() *ComponentBase { return d.base }

// SetOnClose registers a callback invoked when the user presses Escape.
// The application root typically wires this to nav.Pages.Pop.
func (d *DebugOverlay) SetOnClose(fn func()) *DebugOverlay {
	d.onClose = fn
	return d
}

// Name implements nav.Component.
func (d *DebugOverlay) Name() string { return d.base.Name() }

// Start implements nav.Component lifecycle.
func (d *DebugOverlay) Start() { d.base.Start() }

// Stop implements nav.Component lifecycle.
func (d *DebugOverlay) Stop() { d.base.Stop() }

// Hints implements nav.Component.
func (d *DebugOverlay) Hints() []KeyHint { return d.base.Hints() }

func (d *DebugOverlay) onStart() {
	// Enable the bus while the overlay is visible; restore on Stop.
	d.prevEnabled = bus.Enabled()
	bus.SetEnabled(true)

	// Seed with recent ring contents.
	for _, e := range bus.Default().Recent(d.cap) {
		d.append(e)
	}
	d.refresh()

	// Subscribe; cleanup is automatic via Subs() on Stop.
	unsub := bus.Default().Subscribe(nil, func(e bus.Event) {
		d.mu.Lock()
		paused := d.paused
		d.mu.Unlock()
		if paused {
			return
		}
		d.append(e)
		theme.QueueUpdateDraw(d.refresh)
	})
	d.base.Subs().Add(unsub)
}

func (d *DebugOverlay) onStop() {
	bus.SetEnabled(d.prevEnabled)
}

func (d *DebugOverlay) append(e bus.Event) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.events = append(d.events, e)
	if len(d.events) > d.cap {
		d.events = d.events[len(d.events)-d.cap:]
	}
}

func (d *DebugOverlay) renderHeader() {
	d.Table.SetCell(0, 0, tview.NewTableCell("seq").SetSelectable(false).SetAttributes(tcell.AttrBold))
	d.Table.SetCell(0, 1, tview.NewTableCell("time").SetSelectable(false).SetAttributes(tcell.AttrBold))
	d.Table.SetCell(0, 2, tview.NewTableCell("src").SetSelectable(false).SetAttributes(tcell.AttrBold))
	d.Table.SetCell(0, 3, tview.NewTableCell("kind").SetSelectable(false).SetAttributes(tcell.AttrBold))
	d.Table.SetCell(0, 4, tview.NewTableCell("payload").SetSelectable(false).SetAttributes(tcell.AttrBold))
}

func (d *DebugOverlay) refresh() {
	d.mu.Lock()
	events := make([]bus.Event, len(d.events))
	copy(events, d.events)
	filter := make(map[string]bool, len(d.sources))
	for k, v := range d.sources {
		filter[k] = v
	}
	paused := d.paused
	d.mu.Unlock()

	d.Table.Clear()
	d.renderHeader()
	title := " Bus Debug "
	if paused {
		title = " Bus Debug (paused) "
	}
	if len(filter) > 0 {
		keys := make([]string, 0, len(filter))
		for k := range filter {
			keys = append(keys, k)
		}
		title = strings.TrimRight(title, " ") + " [" + strings.Join(keys, ",") + "] "
	}
	d.Table.SetTitle(title)

	row := 1
	for i := len(events) - 1; i >= 0 && row <= d.cap; i-- {
		e := events[i]
		if len(filter) > 0 && !filter[e.Source] {
			continue
		}
		d.Table.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", e.Seq)))
		d.Table.SetCell(row, 1, tview.NewTableCell(e.Time.Format("15:04:05.000")))
		d.Table.SetCell(row, 2, tview.NewTableCell(e.Source))
		d.Table.SetCell(row, 3, tview.NewTableCell(e.Kind))
		d.Table.SetCell(row, 4, tview.NewTableCell(formatPayload(e.Payload)))
		row++
	}
}

func formatPayload(p any) string {
	if p == nil {
		return ""
	}
	s := fmt.Sprintf("%+v", p)
	if len(s) > 120 {
		s = s[:117] + "..."
	}
	return s
}

func (d *DebugOverlay) handleKey(event *tcell.EventKey, _ func(tview.Primitive)) bool {
	if event.Key() == tcell.KeyEscape {
		if d.onClose != nil {
			d.onClose()
		}
		return true
	}
	switch event.Rune() {
	case 'b':
		d.toggleSource(bus.SourceBinding)
	case 't':
		d.toggleSource(bus.SourceTheme)
	case 'n':
		d.toggleSource(bus.SourceNav)
	case 'a':
		d.toggleSource(bus.SourceAsync)
	case 'i':
		d.toggleSource(bus.SourceInput)
	case 'e':
		d.toggleSource(bus.SourceEffect)
	case 'c':
		d.mu.Lock()
		d.sources = map[string]bool{}
		d.mu.Unlock()
	case 'p':
		d.mu.Lock()
		d.paused = !d.paused
		d.mu.Unlock()
	case 'x':
		d.mu.Lock()
		d.events = nil
		d.mu.Unlock()
	default:
		return false
	}
	d.refresh()
	return true
}

func (d *DebugOverlay) toggleSource(src string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sources[src] {
		delete(d.sources, src)
	} else {
		d.sources[src] = true
	}
}


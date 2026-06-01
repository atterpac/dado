package components

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/bus"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// eventLogTool is the panel tool that tails the bus event ring in a table.
// It is the toolbar-tool form of the former full-screen DebugOverlay: same
// source filters (b/t/n/a/i/e), pause (p), and clear (x).
type eventLogTool struct {
	table *core.Table

	mu          sync.Mutex
	events      []bus.Event // newest last
	cap         int
	sources     map[string]bool // empty = all
	paused      bool
	prevEnabled bool
	unsub       func()
}

// NewEventLogTool builds the event-log panel tool. capacity bounds the visible
// event list (defaults to 500 if <= 0).
func NewEventLogTool(capacity int) DebugTool {
	if capacity <= 0 {
		capacity = 500
	}
	tbl := core.NewTable()
	tbl.SetSelectable(true, false)
	tbl.SetBorder(true)
	tbl.SetTitle(" Bus Events ")

	t := &eventLogTool{
		table:   tbl,
		cap:     capacity,
		sources: map[string]bool{},
	}
	t.renderHeader()
	return t
}

func (t *eventLogTool) Name() string        { return "Events" }
func (t *eventLogTool) Kind() DebugToolKind { return DebugToolPanel }

func (t *eventLogTool) Hints() []KeyHint {
	return []KeyHint{
		{Key: "b/t/n/a/i/e", Description: "Filter source"},
		{Key: "c", Description: "Clear filter"},
		{Key: "p", Description: "Pause"},
		{Key: "x", Description: "Clear"},
		{Key: "↑/↓", Description: "Scroll"},
	}
}

func (t *eventLogTool) Activate() {
	t.table.Focus()
	t.prevEnabled = bus.Enabled()
	bus.SetEnabled(true)

	for _, e := range bus.Default().Recent(t.cap) {
		t.append(e)
	}
	t.refresh()

	t.unsub = bus.Default().Subscribe(nil, func(e bus.Event) {
		t.mu.Lock()
		paused := t.paused
		t.mu.Unlock()
		if paused {
			return
		}
		t.append(e)
		theme.QueueUpdateDraw(t.refresh)
	})
}

func (t *eventLogTool) Deactivate() {
	if t.unsub != nil {
		t.unsub()
		t.unsub = nil
	}
	t.table.Blur()
	bus.SetEnabled(t.prevEnabled)
}

func (t *eventLogTool) Draw(screen tcell.Screen, x, y, w, h int) {
	t.table.SetRect(x, y, w, h)
	t.table.Draw(screen)
}

func (t *eventLogTool) HandleKey(ev *tcell.EventKey) bool {
	switch ev.Rune() {
	case 'b':
		t.toggleSource(bus.SourceBinding)
	case 't':
		t.toggleSource(bus.SourceTheme)
	case 'n':
		t.toggleSource(bus.SourceNav)
	case 'a':
		t.toggleSource(bus.SourceAsync)
	case 'i':
		t.toggleSource(bus.SourceInput)
	case 'e':
		t.toggleSource(bus.SourceEffect)
	case 'c':
		t.mu.Lock()
		t.sources = map[string]bool{}
		t.mu.Unlock()
	case 'p':
		t.mu.Lock()
		t.paused = !t.paused
		t.mu.Unlock()
	case 'x':
		t.mu.Lock()
		t.events = nil
		t.mu.Unlock()
	default:
		// Delegate navigation (arrows/home/end) to the table.
		return t.table.HandleKey(ev)
	}
	t.refresh()
	return true
}

func (t *eventLogTool) append(e bus.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)
	if len(t.events) > t.cap {
		t.events = t.events[len(t.events)-t.cap:]
	}
}

func (t *eventLogTool) renderHeader() {
	t.table.SetCell(0, 0, core.NewTableCell("seq"))
	t.table.SetCell(0, 1, core.NewTableCell("time"))
	t.table.SetCell(0, 2, core.NewTableCell("src"))
	t.table.SetCell(0, 3, core.NewTableCell("kind"))
	t.table.SetCell(0, 4, core.NewTableCell("payload"))
}

func (t *eventLogTool) refresh() {
	t.mu.Lock()
	events := make([]bus.Event, len(t.events))
	copy(events, t.events)
	filter := make(map[string]bool, len(t.sources))
	for k, v := range t.sources {
		filter[k] = v
	}
	paused := t.paused
	t.mu.Unlock()

	t.table.Clear()
	t.renderHeader()
	title := " Bus Events "
	if paused {
		title = " Bus Events (paused) "
	}
	if len(filter) > 0 {
		keys := make([]string, 0, len(filter))
		for k := range filter {
			keys = append(keys, k)
		}
		title = strings.TrimRight(title, " ") + " [" + strings.Join(keys, ",") + "] "
	}
	t.table.SetTitle(title)

	row := 1
	for i := len(events) - 1; i >= 0 && row <= t.cap; i-- {
		e := events[i]
		if len(filter) > 0 && !filter[e.Source] {
			continue
		}
		t.table.SetCell(row, 0, core.NewTableCell(fmt.Sprintf("%d", e.Seq)))
		t.table.SetCell(row, 1, core.NewTableCell(e.Time.Format("15:04:05.000")))
		t.table.SetCell(row, 2, core.NewTableCell(e.Source))
		t.table.SetCell(row, 3, core.NewTableCell(e.Kind))
		t.table.SetCell(row, 4, core.NewTableCell(formatDebugPayload(e.Payload)))
		row++
	}
}

func (t *eventLogTool) toggleSource(src string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sources[src] {
		delete(t.sources, src)
	} else {
		t.sources[src] = true
	}
}

func formatDebugPayload(p any) string {
	if p == nil {
		return ""
	}
	s := fmt.Sprintf("%+v", p)
	if len(s) > 120 {
		s = s[:117] + "..."
	}
	return s
}

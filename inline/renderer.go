package inline

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
)

var (
	// ErrRendererClosed is returned when an operation is attempted after Close.
	ErrRendererClosed = errors.New("inline: renderer is closed")
	// ErrNilFrame is returned when Render is called with a nil frame.
	ErrNilFrame = errors.New("inline: nil frame")
)

const (
	ansiReset      = "\x1b[0m"
	ansiHideCursor = "\x1b[?25l"
	ansiShowCursor = "\x1b[?25h"
	ansiEraseLine  = "\x1b[2K"
	ansiSyncStart  = "\x1b[?2026h"
	ansiSyncEnd    = "\x1b[?2026l"
)

// RendererOption configures a Renderer.
type RendererOption func(*Renderer)

// WithOutput sets the renderer's destination.
func WithOutput(w io.Writer) RendererOption {
	return func(r *Renderer) {
		if w != nil {
			r.out = w
		}
	}
}

// WithTerminalOutput overrides terminal detection. This is useful for remote
// PTYs and deterministic tests whose writer is not an *os.File.
func WithTerminalOutput(terminal bool) RendererOption {
	return func(r *Renderer) {
		r.terminal = terminal
		r.terminalSet = true
	}
}

// WithSynchronizedOutput wraps terminal updates in DEC mode 2026. Enable this
// only when the destination terminal is known to support synchronized output.
func WithSynchronizedOutput(enabled bool) RendererOption {
	return func(r *Renderer) { r.synchronized = enabled }
}

// Renderer owns a dynamic region in the terminal's normal screen buffer. It
// leaves scrollback intact and serializes all writes, including persistent
// lines inserted above the live frame.
//
// Unlike tcell.Screen, Renderer never enters the alternate screen, clears the
// whole terminal, or reads input. Close should be called to restore cursor
// visibility and, for non-terminal output, emit the final frame.
type Renderer struct {
	mu sync.Mutex

	out          io.Writer
	terminal     bool
	terminalSet  bool
	synchronized bool
	closed       bool

	last *frameSnapshot
}

// NewRenderer constructs an independent inline renderer. By default it writes
// to standard output and detects whether that writer is a terminal.
func NewRenderer(options ...RendererOption) *Renderer {
	r := &Renderer{out: os.Stdout}
	for _, option := range options {
		if option != nil {
			option(r)
		}
	}
	if !r.terminalSet {
		r.terminal = writerIsTerminal(r.out)
	}
	return r
}

// Render updates the renderer's managed region to match frame. Terminal
// output redraws only changed rows. Non-terminal output retains only the most
// recent frame and emits it once during Close, avoiding animation noise in
// redirected logs.
func (r *Renderer) Render(frame *Frame) error {
	if frame == nil {
		return ErrNilFrame
	}
	next := snapshotFrame(frame)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRendererClosed
	}
	if snapshotsEqual(r.last, next) {
		return nil
	}
	if !r.terminal {
		r.last = next
		return nil
	}

	var buf bytes.Buffer
	r.beginUpdate(&buf)
	r.moveCursorToBaseline(&buf)
	r.renderSnapshot(&buf, next, false)
	r.positionCursor(&buf, next)
	r.endUpdate(&buf)
	if err := writeAll(r.out, buf.Bytes()); err != nil {
		return fmt.Errorf("inline: render: %w", err)
	}
	r.last = next
	return nil
}

// Println writes a persistent line above the managed frame. Embedded newlines
// create multiple persistent lines. ANSI styling is passed through unchanged.
func (r *Renderer) Println(args ...any) error {
	return r.printPersistent(fmt.Sprint(args...))
}

// Printf formats and writes persistent output above the managed frame.
func (r *Renderer) Printf(format string, args ...any) error {
	return r.printPersistent(fmt.Sprintf(format, args...))
}

func (r *Renderer) printPersistent(text string) error {
	lines := splitLines(text)

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRendererClosed
	}
	if !r.terminal {
		for _, line := range lines {
			if _, err := fmt.Fprintln(r.out, stripANSI(line)); err != nil {
				return fmt.Errorf("inline: print persistent output: %w", err)
			}
		}
		return nil
	}

	var buf bytes.Buffer
	r.beginUpdate(&buf)
	r.moveCursorToBaseline(&buf)
	if r.last != nil && r.last.height > 0 {
		cursorUp(&buf, r.last.height)
	}
	for _, line := range lines {
		buf.WriteByte('\r')
		buf.WriteString(ansiEraseLine)
		buf.WriteString(line)
		buf.WriteString(ansiReset)
		buf.WriteString("\r\n")
	}
	if r.last != nil {
		r.renderRowsAtCurrentPosition(&buf, r.last)
	}
	r.positionCursor(&buf, r.last)
	r.endUpdate(&buf)
	if err := writeAll(r.out, buf.Bytes()); err != nil {
		return fmt.Errorf("inline: print persistent output: %w", err)
	}
	return nil
}

// Clear erases the managed region and leaves the cursor at its first row.
// Persistent output previously inserted above the region is not removed.
func (r *Renderer) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRendererClosed
	}
	if r.last == nil {
		return nil
	}
	if !r.terminal {
		r.last = nil
		return nil
	}

	var buf bytes.Buffer
	r.beginUpdate(&buf)
	r.moveCursorToBaseline(&buf)
	if r.last.height > 0 {
		cursorUp(&buf, r.last.height)
		for range r.last.height {
			buf.WriteByte('\r')
			buf.WriteString(ansiEraseLine)
			buf.WriteString("\r\n")
		}
		cursorUp(&buf, r.last.height)
	}
	buf.WriteString(ansiReset)
	buf.WriteString(ansiShowCursor)
	r.endUpdate(&buf)
	if err := writeAll(r.out, buf.Bytes()); err != nil {
		return fmt.Errorf("inline: clear: %w", err)
	}
	r.last = nil
	return nil
}

// Close restores the cursor and releases the renderer. The final terminal
// frame remains visible. For non-terminal destinations, Close emits the final
// frame once as plain text.
func (r *Renderer) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}

	var err error
	if r.terminal {
		var buf bytes.Buffer
		r.moveCursorToBaseline(&buf)
		buf.WriteString(ansiReset)
		buf.WriteString(ansiShowCursor)
		err = writeAll(r.out, buf.Bytes())
	} else if r.last != nil {
		err = writeAll(r.out, []byte(plainSnapshot(r.last)))
	}
	r.closed = true
	if err != nil {
		return fmt.Errorf("inline: close: %w", err)
	}
	return nil
}

func (r *Renderer) beginUpdate(buf *bytes.Buffer) {
	if r.synchronized {
		buf.WriteString(ansiSyncStart)
	}
	buf.WriteString(ansiHideCursor)
}

func (r *Renderer) endUpdate(buf *bytes.Buffer) {
	if r.synchronized {
		buf.WriteString(ansiSyncEnd)
	}
}

// moveCursorToBaseline restores the renderer invariant that the cursor is in
// column zero on the line immediately following the managed region.
func (r *Renderer) moveCursorToBaseline(buf *bytes.Buffer) {
	if r.last == nil || r.last.cursor == nil {
		return
	}
	buf.WriteByte('\r')
	cursorDown(buf, r.last.height-r.last.cursor.Y)
}

func (r *Renderer) renderSnapshot(buf *bytes.Buffer, next *frameSnapshot, force bool) {
	oldHeight := 0
	if r.last != nil {
		oldHeight = r.last.height
		if oldHeight > 0 {
			cursorUp(buf, oldHeight)
		}
	}

	rows := max(oldHeight, next.height)
	for y := range rows {
		changed := force || r.last == nil || y >= oldHeight || y >= next.height
		if !changed {
			changed = !rowsEqual(r.last.rows[y], next.rows[y])
		}
		if changed {
			buf.WriteByte('\r')
			buf.WriteString(ansiEraseLine)
			if y < next.height {
				writeStyledRow(buf, next.rows[y])
			}
		}
		buf.WriteString("\r\n")
	}
	if next.height < oldHeight {
		cursorUp(buf, oldHeight-next.height)
	}
}

func (r *Renderer) renderRowsAtCurrentPosition(buf *bytes.Buffer, frame *frameSnapshot) {
	for _, row := range frame.rows {
		buf.WriteByte('\r')
		buf.WriteString(ansiEraseLine)
		writeStyledRow(buf, row)
		buf.WriteString("\r\n")
	}
}

func (r *Renderer) positionCursor(buf *bytes.Buffer, frame *frameSnapshot) {
	if frame == nil || frame.cursor == nil {
		buf.WriteString(ansiHideCursor)
		return
	}
	cursorUp(buf, frame.height-frame.cursor.Y)
	cursorColumn(buf, frame.cursor.X+1)
	buf.WriteString(ansiShowCursor)
}

type frameCell struct {
	text  string
	style tcell.Style
	width int
}

type frameSnapshot struct {
	width  int
	height int
	rows   [][]frameCell
	cursor *Cursor
}

func snapshotFrame(frame *Frame) *frameSnapshot {
	s := &frameSnapshot{
		width:  frame.width,
		height: frame.height,
		rows:   make([][]frameCell, frame.height),
	}
	if frame.cursor != nil {
		cursor := *frame.cursor
		s.cursor = &cursor
	}
	for y := range frame.height {
		s.rows[y] = make([]frameCell, frame.width)
		for x := range frame.width {
			text, style, width := frame.cells.Get(x, y)
			s.rows[y][x] = frameCell{text: text, style: style, width: width}
		}
	}
	return s
}

func snapshotsEqual(a, b *frameSnapshot) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.width != b.width || a.height != b.height || !cursorsEqual(a.cursor, b.cursor) {
		return false
	}
	for y := range a.height {
		if !rowsEqual(a.rows[y], b.rows[y]) {
			return false
		}
	}
	return true
}

func cursorsEqual(a, b *Cursor) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func rowsEqual(a, b []frameCell) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func plainSnapshot(frame *frameSnapshot) string {
	var buf strings.Builder
	for _, row := range frame.rows {
		last := lastVisibleCell(row)
		for x := 0; x < last; {
			cell := row[x]
			if cell.width < 1 {
				cell.width = 1
			}
			if x+cell.width > len(row) {
				buf.WriteByte(' ')
			} else {
				buf.WriteString(cell.text)
			}
			x += cell.width
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	return strings.Split(text, "\n")
}

func writerIsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

func cursorUp(buf *bytes.Buffer, count int) {
	if count > 0 {
		fmt.Fprintf(buf, "\x1b[%dA", count)
	}
}

func cursorDown(buf *bytes.Buffer, count int) {
	if count > 0 {
		fmt.Fprintf(buf, "\x1b[%dB", count)
	}
}

func cursorColumn(buf *bytes.Buffer, column int) {
	if column < 1 {
		column = 1
	}
	fmt.Fprintf(buf, "\x1b[%dG", column)
}

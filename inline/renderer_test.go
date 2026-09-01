package inline

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRendererInitialRenderAndNoop(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(true))
	frame := textFrame(5, "one", "two")

	require.NoError(t, renderer.Render(frame))
	assert.Equal(t,
		ansiHideCursor+
			"\r"+ansiEraseLine+ansiReset+"one"+ansiReset+"\r\n"+
			"\r"+ansiEraseLine+ansiReset+"two"+ansiReset+"\r\n"+
			ansiHideCursor,
		output.String(),
	)

	output.Reset()
	require.NoError(t, renderer.Render(frame))
	assert.Empty(t, output.String(), "an unchanged frame should not be flushed")
}

func TestRendererRedrawsOnlyChangedRows(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(true))
	require.NoError(t, renderer.Render(textFrame(5, "one", "two")))
	output.Reset()

	require.NoError(t, renderer.Render(textFrame(5, "one", "TWO")))
	assert.Equal(t,
		ansiHideCursor+"\x1b[2A"+
			"\r\n"+
			"\r"+ansiEraseLine+ansiReset+"TWO"+ansiReset+"\r\n"+
			ansiHideCursor,
		output.String(),
	)
}

func TestRendererClearsRowsWhenFrameShrinks(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(true))
	require.NoError(t, renderer.Render(textFrame(5, "one", "two", "three")))
	output.Reset()

	require.NoError(t, renderer.Render(textFrame(5, "one")))
	assert.Equal(t,
		ansiHideCursor+"\x1b[3A"+
			"\r\n"+
			"\r"+ansiEraseLine+"\r\n"+
			"\r"+ansiEraseLine+"\r\n"+
			"\x1b[2A"+ansiHideCursor,
		output.String(),
	)
}

func TestRendererPrintlnInsertsAboveAndRedrawsFrame(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(true))
	require.NoError(t, renderer.Render(textFrame(5, "one", "two")))
	output.Reset()

	require.NoError(t, renderer.Println("built"))
	assert.Equal(t,
		ansiHideCursor+"\x1b[2A"+
			"\r"+ansiEraseLine+"built"+ansiReset+"\r\n"+
			"\r"+ansiEraseLine+ansiReset+"one"+ansiReset+"\r\n"+
			"\r"+ansiEraseLine+ansiReset+"two"+ansiReset+"\r\n"+
			ansiHideCursor,
		output.String(),
	)
}

func TestRendererTracksVisibleCursorAndRestoresBaseline(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(true))
	frame := textFrame(6, "hello", "world")
	frame.ShowCursor(2, 1)
	require.NoError(t, renderer.Render(frame))
	assert.True(t, strings.HasSuffix(output.String(), "\x1b[1A\x1b[3G"+ansiShowCursor))

	output.Reset()
	require.NoError(t, renderer.Close())
	assert.Equal(t, "\r\x1b[1B"+ansiReset+ansiShowCursor, output.String())
}

func TestRendererClearErasesManagedRegion(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(true))
	require.NoError(t, renderer.Render(textFrame(4, "a", "b")))
	output.Reset()

	require.NoError(t, renderer.Clear())
	assert.Equal(t,
		ansiHideCursor+"\x1b[2A"+
			"\r"+ansiEraseLine+"\r\n"+
			"\r"+ansiEraseLine+"\r\n"+
			"\x1b[2A"+ansiReset+ansiShowCursor,
		output.String(),
	)
}

func TestRendererNonTerminalEmitsOnlyFinalPlainFrame(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(false))
	require.NoError(t, renderer.Render(textFrame(8, "working")))
	require.NoError(t, renderer.Println("downloaded"))
	require.NoError(t, renderer.Render(textFrame(8, "done")))
	assert.Equal(t, "downloaded\n", output.String())

	require.NoError(t, renderer.Close())
	assert.Equal(t, "downloaded\ndone\n", output.String())
	require.NoError(t, renderer.Close(), "Close should be idempotent")
}

func TestRendererStylesAndWideLastColumn(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(true))
	frame := NewFrame(3, 1)
	style := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.NewRGBColor(1, 2, 3)).Bold(true)
	frame.DrawString(0, 0, "A", style)
	frame.Put(2, 0, "界", tcell.StyleDefault)

	require.NoError(t, renderer.Render(frame))
	assert.Contains(t, output.String(), "\x1b[1;91;48;2;1;2;3mA")
	assert.NotContains(t, output.String(), "界", "a wide grapheme in the final column must not trigger wrapping")
}

func TestRendererSynchronizedOutput(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := NewRenderer(
		WithOutput(&output),
		WithTerminalOutput(true),
		WithSynchronizedOutput(true),
	)
	require.NoError(t, renderer.Render(textFrame(3, "ok")))
	assert.True(t, strings.HasPrefix(output.String(), ansiSyncStart+ansiHideCursor))
	assert.True(t, strings.HasSuffix(output.String(), ansiHideCursor+ansiSyncEnd))
}

func TestRendererErrors(t *testing.T) {
	t.Parallel()

	renderer := NewRenderer(WithOutput(io.Discard), WithTerminalOutput(false))
	assert.ErrorIs(t, renderer.Render(nil), ErrNilFrame)
	require.NoError(t, renderer.Close())
	assert.ErrorIs(t, renderer.Render(NewFrame(1, 1)), ErrRendererClosed)
	assert.ErrorIs(t, renderer.Println("late"), ErrRendererClosed)
}

func TestRendererSerializesConcurrentWrites(t *testing.T) {
	t.Parallel()

	output := &lockedBuffer{}
	renderer := NewRenderer(WithOutput(output), WithTerminalOutput(true))
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			require.NoError(t, renderer.Render(textFrame(8, "frame")))
		}()
		go func(i int) {
			defer wg.Done()
			require.NoError(t, renderer.Printf("log %d", i))
		}(i)
	}
	wg.Wait()
	require.NoError(t, renderer.Close())
	assert.NotEmpty(t, output.String())
}

func TestRendererPropagatesWriteErrors(t *testing.T) {
	t.Parallel()

	want := errors.New("boom")
	renderer := NewRenderer(WithOutput(errorWriter{err: want}), WithTerminalOutput(true))
	assert.ErrorIs(t, renderer.Render(textFrame(2, "x")), want)
}

func textFrame(width int, lines ...string) *Frame {
	frame := NewFrame(width, len(lines))
	for y, line := range lines {
		frame.DrawString(0, y, line, tcell.StyleDefault)
	}
	return frame
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

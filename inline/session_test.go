package inline

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionRunsFormFromKeyStream(t *testing.T) {
	t.Parallel()
	form := NewForm("Create").Add(
		NewTextField("name", "Name").Required(),
		NewSelectField("channel", "Channel", NewChoice("stable", "Stable"), NewChoice("preview", "Preview")).Required(),
		NewMultiSelectField("targets", "Targets", NewChoice("linux", "Linux"), NewChoice("darwin", "macOS")).MinSelected(1),
	)
	input := bytes.NewBufferString("demo\r\x1b[B\r \r")
	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(false))
	session := NewSession(renderer, WithSessionInput(input), WithSessionWidth(48))
	result, err := session.Run(context.Background(), form)
	require.NoError(t, err)
	assert.Equal(t, "demo", result["name"])
	assert.Equal(t, "preview", result["channel"])
	assert.Equal(t, []string{"linux"}, result["targets"])
	require.NoError(t, renderer.Close())
	assert.Contains(t, output.String(), "Name")
	assert.Contains(t, output.String(), "demo")
}

func TestSessionCancellation(t *testing.T) {
	t.Parallel()
	form := NewForm("Cancel").Add(NewTextField("name", "Name"))
	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(false))
	session := NewSession(renderer, WithSessionInput(bytes.NewReader([]byte{3})), WithSessionWidth(30))
	_, err := session.Run(context.Background(), form)
	assert.ErrorIs(t, err, ErrFormCancelled)
}

func TestSessionRejectsNilValues(t *testing.T) {
	t.Parallel()
	session := NewSession(nil, WithSessionInput(bytes.NewReader(nil)))
	_, err := session.Run(context.Background(), NewForm(""))
	assert.Error(t, err)
}

func TestSessionReadsFragmentedEscapeSequence(t *testing.T) {
	t.Parallel()
	form := NewForm("Choose").Add(
		NewSelectField("channel", "Channel", NewChoice("stable", "Stable"), NewChoice("preview", "Preview")),
	)
	input := &oneByteReader{data: []byte("\x1b[B\r")}
	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(false))
	result, err := NewSession(renderer, WithSessionInput(input)).Run(context.Background(), form)
	require.NoError(t, err)
	assert.Equal(t, "preview", result["channel"])
}

func TestSessionReadsFragmentedUTF8(t *testing.T) {
	t.Parallel()
	form := NewForm("Name").Add(NewTextField("name", "Name"))
	input := &oneByteReader{data: []byte("é\r")}
	var output bytes.Buffer
	renderer := NewRenderer(WithOutput(&output), WithTerminalOutput(false))
	result, err := NewSession(renderer, WithSessionInput(input)).Run(context.Background(), form)
	require.NoError(t, err)
	assert.Equal(t, "é", result["name"])
}

func TestSessionContextCancellationInterruptsTerminalRead(t *testing.T) {
	t.Parallel()
	input, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = input.Close()
		_ = writer.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	renderer := NewRenderer(WithOutput(io.Discard), WithTerminalOutput(false))
	started := time.Now()
	_, err = NewSession(renderer, WithSessionInput(input)).Run(ctx, NewForm("Wait").Add(NewTextField("name", "Name")))
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(started), 250*time.Millisecond)
}

type oneByteReader struct{ data []byte }

func (r *oneByteReader) Read(target []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	target[0] = r.data[0]
	r.data = r.data[1:]
	return 1, nil
}

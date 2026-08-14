package inline

import (
	"bytes"
	"context"
	"testing"

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

package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/dado/core"
)

// TestMasterDetailView_HandleMouse_ForwardsToMaster verifies a click in the
// master pane reaches the embedded table and moves its selection. The view must
// be drawn first so the internal Split/Panel rects are established.
func TestMasterDetailView_HandleMouse_ForwardsToMaster(t *testing.T) {
	tbl := NewTable()
	tbl.SetHeaders("Name")
	for i := 0; i < 6; i++ {
		tbl.AddRow("row")
	}

	mdv := NewMasterDetailView().
		SetMasterTitle("Master").
		SetMasterContent(tbl)

	screen := newTestScreen(80, 24)
	mdv.SetRect(0, 0, 80, 24)
	mdv.Draw(screen) // establishes pane/table rects

	// Table lives inside master panel: panel border (+1) and title occupy the
	// top edge, so its first data row is a few cells down from the top-left.
	// Find a Y that maps to a valid data row by probing the table directly.
	tx, ty, _, th := tbl.Table.GetInnerRect()
	require.Positive(t, th)

	// Click the second visible data cell (skip header row at ty).
	clickY := ty + 2 // header at ty, data row 0 at ty+1, data row 1 at ty+2
	consumed, _ := mdv.HandleMouse(core.MouseLeftClick, mouseAt(tx+1, clickY))
	assert.True(t, consumed)
	assert.Equal(t, 1, tbl.SelectedRow())
	assert.True(t, mdv.IsMasterFocused())
}

package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newZoomERD() (*ERDGraph, *ERDTable) {
	cols := []ERDColumn{
		{Name: "id", IsPK: true},
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}
	tbl := &ERDTable{ID: "t", Name: "t", Columns: cols}
	g := NewERDGraph()
	g.SetData([]*ERDTable{tbl}, nil)
	g.SetRect(0, 0, 80, 24)
	return g, tbl
}

func TestERDGraph_Zoom_Levels(t *testing.T) {
	g, tbl := newZoomERD()

	assert.Equal(t, 4, g.MaxDetail())
	assert.Equal(t, 4, g.Detail(), "all columns shown by default")
	assert.Len(t, g.visibleColumns(tbl), 4)
	assert.Equal(t, 3+4, tbl.height)

	// Zoom out to two columns: PK kept first, order preserved.
	g.SetDetail(2)
	assert.Equal(t, 2, g.Detail())
	vis := g.visibleColumns(tbl)
	assert.Len(t, vis, 2)
	assert.True(t, vis[0].IsPK, "key column prioritized when truncating")
	assert.Equal(t, 3+2, tbl.height)

	// Fully zoomed out → name-only compact box.
	g.SetDetail(0)
	assert.Equal(t, 0, g.Detail())
	assert.Empty(t, g.visibleColumns(tbl))
	assert.Equal(t, 3, tbl.height)

	// Zooming in past the max clamps to "all".
	g.SetDetail(99)
	assert.Equal(t, 4, g.Detail())
	assert.Len(t, g.visibleColumns(tbl), 4)
}

func TestERDGraph_Zoom_Steppers(t *testing.T) {
	g, _ := newZoomERD() // starts at "all" (4)

	g.ZoomOut()
	assert.Equal(t, 3, g.Detail())
	g.ZoomOut()
	assert.Equal(t, 2, g.Detail())

	g.ZoomIn()
	assert.Equal(t, 3, g.Detail())

	// Clamp at the floor.
	g.SetDetail(0)
	g.ZoomOut()
	assert.Equal(t, 0, g.Detail())

	// Clamp at the ceiling ("all" columns).
	g.SetDetail(g.MaxDetail())
	g.ZoomIn()
	assert.Equal(t, 4, g.Detail())
}

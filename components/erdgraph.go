package components

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// ERDCardinality describes the cardinality of a relationship.
type ERDCardinality int

const (
	OneToOne   ERDCardinality = iota
	OneToMany
	ManyToMany
)

// ERDRelationType describes the visual style of a relationship line.
type ERDRelationType int

const (
	ERDSolid  ERDRelationType = iota // Enforced FK
	ERDDashed                        // Logical/inferred
)

// ERDColumn represents a single column in an ERD table node.
type ERDColumn struct {
	Name     string
	Type     string
	IsPK     bool
	IsFK     bool
	FKTarget string // "schema.table.column" or "table.column"
}

// ERDTable represents a table node in the ERD.
type ERDTable struct {
	ID      string
	Name    string
	Columns []ERDColumn
	Data    any // arbitrary user data

	// Layout (computed by layout algorithm)
	x, y          int
	width, height int
}

// ERDRelation represents a foreign key relationship between two tables.
type ERDRelation struct {
	FromTable  string
	FromColumn string
	ToTable    string
	ToColumn   string

	Cardinality ERDCardinality
	Type        ERDRelationType
}

// erdGraphData holds all ERD data for an ERDGraph.
type erdGraphData struct {
	tables     map[string]*ERDTable
	tableOrder []string // deterministic iteration order
	relations  []*ERDRelation
	focusID    string
}

// ERDGraph is a 2D ERD visualization component that shows database tables
// with their columns and foreign key relationships.
type ERDGraph struct {
	*tview.Box

	data       *erdGraphData
	nodeWidth  int // minimum node width (auto-expanded per table)
	hSpacing   int // horizontal spacing between grid cells
	vSpacing   int // vertical spacing between grid cells

	// Viewport offset for panning
	offsetX, offsetY int

	// Auto-center flag
	needsCenter bool

	// Focus state
	focused bool

	// Callbacks
	onSelect func(table *ERDTable)

	subs Subscriptions
}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (e *ERDGraph) Subs() *Subscriptions { return &e.subs }

// NewERDGraph creates a new ERDGraph component.
func NewERDGraph() *ERDGraph {
	box := tview.NewBox()
	box.SetBackgroundColor(theme.Bg())

	g := &ERDGraph{
		Box:       box,
		nodeWidth: 30,
		hSpacing:  4,
		vSpacing:  2,
	}
	g.subs.Add(theme.Register(box))
	return g
}

// SetData replaces the graph data with the given tables and relations,
// recomputes the layout, and schedules an auto-center on the next draw.
func (g *ERDGraph) SetData(tables []*ERDTable, relations []*ERDRelation) *ERDGraph {
	d := &erdGraphData{
		tables:    make(map[string]*ERDTable, len(tables)),
		relations: relations,
	}
	for _, t := range tables {
		d.tables[t.ID] = t
		d.tableOrder = append(d.tableOrder, t.ID)
	}

	// Default focus to first table
	if len(d.tableOrder) > 0 {
		d.focusID = d.tableOrder[0]
	}

	g.data = d
	g.computeLayout()
	g.needsCenter = true
	return g
}

// SetNodeWidth sets the minimum node width.
func (g *ERDGraph) SetNodeWidth(width int) *ERDGraph {
	g.nodeWidth = width
	return g
}

// SetSpacing sets horizontal and vertical spacing between grid cells.
func (g *ERDGraph) SetSpacing(h, v int) *ERDGraph {
	g.hSpacing = h
	g.vSpacing = v
	return g
}

// SetOnSelect sets the callback fired when the user presses Enter on a focused table.
func (g *ERDGraph) SetOnSelect(fn func(table *ERDTable)) *ERDGraph {
	g.onSelect = fn
	return g
}

// FocusedTable returns the currently focused table, or nil.
func (g *ERDGraph) FocusedTable() *ERDTable {
	if g.data == nil || g.data.focusID == "" {
		return nil
	}
	return g.data.tables[g.data.focusID]
}

// TableOrder returns the ordered list of table IDs.
func (g *ERDGraph) TableOrder() []string {
	if g.data == nil {
		return nil
	}
	return g.data.tableOrder
}

// SetFocusedTable changes the focused table and centers the viewport on it.
func (g *ERDGraph) SetFocusedTable(id string) {
	if g.data == nil {
		return
	}
	if _, ok := g.data.tables[id]; !ok {
		return
	}
	g.data.focusID = id
	g.centerOnFocused()
}

// Focus implements tview.Primitive.
func (g *ERDGraph) Focus(delegate func(tview.Primitive)) {
	g.focused = true
	g.Box.Focus(delegate)
}

// Blur implements tview.Primitive.
func (g *ERDGraph) Blur() {
	g.focused = false
	g.Box.Blur()
}

// HasFocus implements tview.Primitive.
func (g *ERDGraph) HasFocus() bool {
	return g.focused
}

// centerOnFocused centers the viewport on the currently focused node.
func (g *ERDGraph) centerOnFocused() {
	if g.data == nil || g.data.focusID == "" {
		return
	}
	node := g.data.tables[g.data.focusID]
	if node == nil {
		return
	}

	_, _, width, height := g.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	g.offsetX = node.x + node.width/2 - width/2
	g.offsetY = node.y + node.height/2 - height/2
}

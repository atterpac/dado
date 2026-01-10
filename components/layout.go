package components

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Layout wraps tview.Flex with a simpler API for common layout patterns.
type Layout struct {
	*tview.Flex
}

// Row creates a horizontal layout (items arranged left to right).
// Items are given equal weight by default.
func Row(items ...tview.Primitive) *Layout {
	flex := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex.SetBackgroundColor(theme.Bg())
	theme.Register(flex)

	for _, item := range items {
		flex.AddItem(item, 0, 1, false)
	}

	// Set focus on first item
	if len(items) > 0 {
		flex.AddItem(nil, 0, 0, true) // dummy to reset focus tracking
	}

	return &Layout{Flex: flex}
}

// Column creates a vertical layout (items arranged top to bottom).
// Items are given equal weight by default.
func Column(items ...tview.Primitive) *Layout {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.Bg())
	theme.Register(flex)

	for _, item := range items {
		flex.AddItem(item, 0, 1, false)
	}

	return &Layout{Flex: flex}
}

// NewLayout creates an empty layout. Use AddItem to add children.
func NewLayout() *Layout {
	flex := tview.NewFlex()
	flex.SetBackgroundColor(theme.Bg())
	theme.Register(flex)

	return &Layout{Flex: flex}
}

// Horizontal sets the layout direction to horizontal (row).
func (l *Layout) Horizontal() *Layout {
	l.Flex.SetDirection(tview.FlexColumn)
	return l
}

// Vertical sets the layout direction to vertical (column).
func (l *Layout) Vertical() *Layout {
	l.Flex.SetDirection(tview.FlexRow)
	return l
}

// Add adds an item with equal weight (flexible sizing).
func (l *Layout) Add(item tview.Primitive) *Layout {
	l.Flex.AddItem(item, 0, 1, false)
	return l
}

// AddFixed adds an item with a fixed size.
func (l *Layout) AddFixed(item tview.Primitive, size int) *Layout {
	l.Flex.AddItem(item, size, 0, false)
	return l
}

// AddWeighted adds an item with a specific weight.
// Higher weight = more space relative to other weighted items.
func (l *Layout) AddWeighted(item tview.Primitive, weight int) *Layout {
	l.Flex.AddItem(item, 0, weight, false)
	return l
}

// AddFocused adds an item that should receive initial focus.
func (l *Layout) AddFocused(item tview.Primitive) *Layout {
	l.Flex.AddItem(item, 0, 1, true)
	return l
}

// AddFixedFocused adds a fixed-size item that should receive initial focus.
func (l *Layout) AddFixedFocused(item tview.Primitive, size int) *Layout {
	l.Flex.AddItem(item, size, 0, true)
	return l
}

// AddSpacer adds an empty spacer that takes up available space.
func (l *Layout) AddSpacer() *Layout {
	l.Flex.AddItem(nil, 0, 1, false)
	return l
}

// AddFixedSpacer adds an empty spacer with a fixed size.
func (l *Layout) AddFixedSpacer(size int) *Layout {
	l.Flex.AddItem(nil, size, 0, false)
	return l
}

// Clear removes all items from the layout.
func (l *Layout) Clear() *Layout {
	l.Flex.Clear()
	return l
}

// AddItem adds an item with full control over sizing and focus.
// This matches tview.Flex.AddItem signature for compatibility.
func (l *Layout) AddItem(item tview.Primitive, fixedSize, proportion int, focus bool) *Layout {
	l.Flex.AddItem(item, fixedSize, proportion, focus)
	return l
}

// Primitive returns the underlying tview.Flex for advanced usage.
func (l *Layout) Primitive() *tview.Flex {
	return l.Flex
}

// RowBuilder provides a fluent API for building row layouts.
type RowBuilder struct {
	layout *Layout
}

// NewRow creates a new row builder.
func NewRow() *RowBuilder {
	return &RowBuilder{
		layout: Row(),
	}
}

// Add adds an item with equal weight.
func (r *RowBuilder) Add(item tview.Primitive) *RowBuilder {
	r.layout.Add(item)
	return r
}

// Fixed adds an item with fixed width.
func (r *RowBuilder) Fixed(item tview.Primitive, width int) *RowBuilder {
	r.layout.AddFixed(item, width)
	return r
}

// Weighted adds an item with specific weight.
func (r *RowBuilder) Weighted(item tview.Primitive, weight int) *RowBuilder {
	r.layout.AddWeighted(item, weight)
	return r
}

// Focused adds an item that should receive focus.
func (r *RowBuilder) Focused(item tview.Primitive) *RowBuilder {
	r.layout.AddFocused(item)
	return r
}

// Spacer adds empty space.
func (r *RowBuilder) Spacer() *RowBuilder {
	r.layout.AddSpacer()
	return r
}

// FixedSpacer adds fixed empty space.
func (r *RowBuilder) FixedSpacer(width int) *RowBuilder {
	r.layout.AddFixedSpacer(width)
	return r
}

// Build returns the completed layout.
func (r *RowBuilder) Build() *Layout {
	return r.layout
}

// ColumnBuilder provides a fluent API for building column layouts.
type ColumnBuilder struct {
	layout *Layout
}

// NewColumn creates a new column builder.
func NewColumn() *ColumnBuilder {
	return &ColumnBuilder{
		layout: Column(),
	}
}

// Add adds an item with equal weight.
func (c *ColumnBuilder) Add(item tview.Primitive) *ColumnBuilder {
	c.layout.Add(item)
	return c
}

// Fixed adds an item with fixed height.
func (c *ColumnBuilder) Fixed(item tview.Primitive, height int) *ColumnBuilder {
	c.layout.AddFixed(item, height)
	return c
}

// Weighted adds an item with specific weight.
func (c *ColumnBuilder) Weighted(item tview.Primitive, weight int) *ColumnBuilder {
	c.layout.AddWeighted(item, weight)
	return c
}

// Focused adds an item that should receive focus.
func (c *ColumnBuilder) Focused(item tview.Primitive) *ColumnBuilder {
	c.layout.AddFocused(item)
	return c
}

// Spacer adds empty space.
func (c *ColumnBuilder) Spacer() *ColumnBuilder {
	c.layout.AddSpacer()
	return c
}

// FixedSpacer adds fixed empty space.
func (c *ColumnBuilder) FixedSpacer(height int) *ColumnBuilder {
	c.layout.AddFixedSpacer(height)
	return c
}

// Build returns the completed layout.
func (c *ColumnBuilder) Build() *Layout {
	return c.layout
}

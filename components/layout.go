package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// Layout wraps core.Flex with a simpler API for common layout patterns.
type Layout struct {
	*core.Flex
	subs Subscriptions
}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (l *Layout) Subs() *Subscriptions { return &l.subs }

// Row creates a horizontal layout (items arranged left to right).
// Items are given equal weight by default.
func Row(items ...core.Widget) *Layout {
	flex := core.NewFlex().SetDirection(core.Row)
	flex.Box.SetBackgroundColor(theme.Bg())
	l := &Layout{Flex: flex}
	l.subs.Add(theme.RegisterFn(func(c tcell.Color) { flex.Box.SetBackgroundColor(c) }))

	for _, item := range items {
		flex.AddItem(item, 0, 1, false)
	}

	return l
}

// Column creates a vertical layout (items arranged top to bottom).
// Items are given equal weight by default.
func Column(items ...core.Widget) *Layout {
	flex := core.NewFlex()
	flex.Box.SetBackgroundColor(theme.Bg())
	l := &Layout{Flex: flex}
	l.subs.Add(theme.RegisterFn(func(c tcell.Color) { flex.Box.SetBackgroundColor(c) }))

	for _, item := range items {
		flex.AddItem(item, 0, 1, false)
	}

	return l
}

// NewLayout creates an empty layout. Use AddItem to add children.
func NewLayout() *Layout {
	flex := core.NewFlex()
	flex.Box.SetBackgroundColor(theme.Bg())
	l := &Layout{Flex: flex}
	l.subs.Add(theme.RegisterFn(func(c tcell.Color) { flex.Box.SetBackgroundColor(c) }))
	return l
}

// Horizontal sets the layout direction to horizontal (row).
func (l *Layout) Horizontal() *Layout {
	l.Flex.SetDirection(core.Row)
	return l
}

// Vertical sets the layout direction to vertical (column).
func (l *Layout) Vertical() *Layout {
	l.Flex.SetDirection(core.Column)
	return l
}

// Add adds an item with equal weight (flexible sizing).
func (l *Layout) Add(item core.Widget) *Layout {
	l.Flex.AddItem(item, 0, 1, false)
	return l
}

// AddFixed adds an item with a fixed size.
func (l *Layout) AddFixed(item core.Widget, size int) *Layout {
	l.Flex.AddItem(item, size, 0, false)
	return l
}

// AddWeighted adds an item with a specific weight.
// Higher weight = more space relative to other weighted items.
func (l *Layout) AddWeighted(item core.Widget, weight int) *Layout {
	l.Flex.AddItem(item, 0, weight, false)
	return l
}

// AddFocused adds an item that should receive initial focus.
func (l *Layout) AddFocused(item core.Widget) *Layout {
	l.Flex.AddItem(item, 0, 1, true)
	return l
}

// AddFixedFocused adds a fixed-size item that should receive initial focus.
func (l *Layout) AddFixedFocused(item core.Widget, size int) *Layout {
	l.Flex.AddItem(item, size, 0, true)
	return l
}

// AddSpacer adds an empty spacer that takes up available space.
func (l *Layout) AddSpacer() *Layout {
	l.Flex.AddItem(new(core.Box), 0, 1, false)
	return l
}

// AddFixedSpacer adds an empty spacer with a fixed size.
func (l *Layout) AddFixedSpacer(size int) *Layout {
	l.Flex.AddItem(new(core.Box), size, 0, false)
	return l
}

// Clear removes all items from the layout.
func (l *Layout) Clear() *Layout {
	l.Flex.Clear()
	return l
}

// AddItem adds an item with full control over sizing and focus.
// This matches core.Flex.AddItem signature for compatibility.
func (l *Layout) AddItem(item core.Widget, fixedSize, proportion int, focus bool) *Layout {
	l.Flex.AddItem(item, fixedSize, proportion, focus)
	return l
}

// Primitive returns the underlying core.Flex for advanced usage.
func (l *Layout) Primitive() *core.Flex {
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
func (r *RowBuilder) Add(item core.Widget) *RowBuilder {
	r.layout.Add(item)
	return r
}

// Fixed adds an item with fixed width.
func (r *RowBuilder) Fixed(item core.Widget, width int) *RowBuilder {
	r.layout.AddFixed(item, width)
	return r
}

// Weighted adds an item with specific weight.
func (r *RowBuilder) Weighted(item core.Widget, weight int) *RowBuilder {
	r.layout.AddWeighted(item, weight)
	return r
}

// Focused adds an item that should receive focus.
func (r *RowBuilder) Focused(item core.Widget) *RowBuilder {
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
func (c *ColumnBuilder) Add(item core.Widget) *ColumnBuilder {
	c.layout.Add(item)
	return c
}

// Fixed adds an item with fixed height.
func (c *ColumnBuilder) Fixed(item core.Widget, height int) *ColumnBuilder {
	c.layout.AddFixed(item, height)
	return c
}

// Weighted adds an item with specific weight.
func (c *ColumnBuilder) Weighted(item core.Widget, weight int) *ColumnBuilder {
	c.layout.AddWeighted(item, weight)
	return c
}

// Focused adds an item that should receive focus.
func (c *ColumnBuilder) Focused(item core.Widget) *ColumnBuilder {
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

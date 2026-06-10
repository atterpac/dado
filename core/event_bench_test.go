package core_test

// Benchmarks for event dispatch: focus changes (callback-slice copy) and mouse
// hit-testing (DescendantsAt tree walk + slice append).
//
// Run:  go test -run=^$ -bench=Focus -benchmem ./core

import (
	"testing"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

func BenchmarkFocusChange(b *testing.B) {
	for _, listeners := range []int{0, 10, 100} {
		fm := core.NewFocusManager()
		for i := 0; i < listeners; i++ {
			fm.OnChange(func(_, _ core.Widget) {})
		}
		a := coretest.NewMockWidget("a")
		w := coretest.NewMockWidget("b")
		b.Run(itoa(listeners)+"listeners", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i&1 == 0 {
					fm.Focus(a)
				} else {
					fm.Focus(w)
				}
			}
		})
	}
}

// BenchmarkDescendantsAt walks a nested Flex tree to hit-test a point, the
// per-mouse-event cost. Depth controls how deep the recursion + slice append go.
func BenchmarkDescendantsAt(b *testing.B) {
	for _, depth := range []int{1, 5, 10} {
		root := core.NewFlex().SetDirection(core.Column)
		cur := root
		for d := 0; d < depth; d++ {
			child := core.NewFlex().SetDirection(core.Column)
			cur.AddItem(child, 0, 1, false)
			cur.AddItem(coretest.NewMockWidget("leaf"), 0, 1, false)
			cur = child
		}
		root.SetRect(0, 0, 160, 48)
		root.Draw(benchScreen(160, 48)) // assign rects down the tree
		b.Run(itoa(depth)+"deep", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = root.DescendantsAt(80, 24)
			}
		})
	}
}

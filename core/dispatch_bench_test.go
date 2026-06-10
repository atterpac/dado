package core

// In-package benchmark for composed key dispatch: App.dispatchKey walks from the
// focused widget up to the root, and each step calls findParent, which itself
// recursively walks the whole tree (calling Container.Children at each node).
// This measures that walk at increasing tree depth.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// benchKeyLeaf is a focusable widget that never consumes keys, forcing
// dispatchKey to walk the entire focus→root chain (worst case).
type benchKeyLeaf struct{ Box }

func (b *benchKeyLeaf) HandleKey(_ *tcell.EventKey) bool { return false }

func BenchmarkDispatchKey(b *testing.B) {
	for _, depth := range []int{1, 5, 10, 20} {
		screen := tcell.NewSimulationScreen("UTF-8")
		_ = screen.Init()
		screen.SetSize(80, 24)
		app := NewAppFromScreen(screen)

		root := NewFlex().SetDirection(Column)
		cur := root
		leaf := &benchKeyLeaf{}
		for d := 0; d < depth; d++ {
			next := NewFlex().SetDirection(Column)
			cur.AddItem(next, 0, 1, false)
			cur.AddItem(&benchKeyLeaf{}, 0, 1, false) // sibling, widens Children()
			if d == depth-1 {
				next.AddItem(leaf, 0, 1, false)
			}
			cur = next
		}
		app.SetRoot(root)
		root.SetRect(0, 0, 80, 24)
		app.focus.Focus(leaf)
		ev := tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)

		b.Run(itoaDepth(depth)+"deep", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				app.dispatchKey(ev)
			}
		})
	}
}

func itoaDepth(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

package core

import "testing"

// appendPathRev underlies dispatchKey: it must return the focused widget first,
// then each ancestor up to root, so keys are offered leaf→root.
func TestAppendPathRev_Order(t *testing.T) {
	root := NewFlex().SetDirection(Column)
	mid := NewFlex().SetDirection(Column)
	leaf := &benchKeyLeaf{}
	sibling := &benchKeyLeaf{}
	mid.AddItem(leaf, 0, 1, false)
	root.AddItem(mid, 0, 1, false)
	root.AddItem(sibling, 0, 1, false)

	path, found := appendPathRev(nil, root, leaf)
	if !found {
		t.Fatal("leaf should be found under root")
	}
	want := []Widget{leaf, mid, root} // leaf-first, then ancestors
	if len(path) != len(want) {
		t.Fatalf("path length: got %d want %d", len(path), len(want))
	}
	for i := range want {
		if path[i] != want[i] {
			t.Errorf("path[%d]: got %p want %p", i, path[i], want[i])
		}
	}
}

func TestAppendPathRev_NotFound(t *testing.T) {
	root := NewFlex()
	root.AddItem(&benchKeyLeaf{}, 0, 1, false)
	orphan := &benchKeyLeaf{}

	path, found := appendPathRev(nil, root, orphan)
	if found {
		t.Fatal("orphan must not be found")
	}
	if len(path) != 0 {
		t.Fatalf("not-found must append nothing, got len %d", len(path))
	}
}

// Reusing the buffer across calls (as dispatchKey does with keyPath[:0]) must
// yield a correct, independent result each time.
func TestAppendPathRev_BufferReuse(t *testing.T) {
	root := NewFlex().SetDirection(Column)
	a := &benchKeyLeaf{}
	b := &benchKeyLeaf{}
	root.AddItem(a, 0, 1, false)
	root.AddItem(b, 0, 1, false)

	var buf []Widget
	buf, _ = appendPathRev(buf[:0], root, a)
	if buf[0] != Widget(a) {
		t.Fatalf("first call: got %p want %p", buf[0], a)
	}
	buf, _ = appendPathRev(buf[:0], root, b)
	if buf[0] != Widget(b) {
		t.Fatalf("reused call: got %p want %p", buf[0], b)
	}
	if len(buf) != 2 {
		t.Fatalf("path len: got %d want 2", len(buf))
	}
}

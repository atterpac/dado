package core

// White-box equivalence test: the arena-based reflow (wrapLineInto) must produce
// exactly the spans/lines the previous ParseTagged+splitWidth implementation did.
// Kept in package core so it can reach the unexported reflow state directly.

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// reference reproduces the pre-optimization wrapping: split on '\n', then per
// line ParseTagged (or a single plain span) followed by splitWidth.
func reference(text string, dynamic bool, width int) []*Text {
	var out []*Text
	for _, raw := range strings.Split(text, "\n") {
		var t *Text
		if dynamic {
			t = ParseTagged(raw, tcell.StyleDefault)
		} else {
			t = NewText().Append(raw, tcell.StyleDefault)
		}
		out = append(out, t.splitWidth(width)...)
	}
	return out
}

func TestReflow_MatchesReference(t *testing.T) {
	cases := []struct {
		name string
		text string
		dyn  bool
	}{
		{"empty", "", false},
		{"plain_short", "hello", false},
		{"plain_exact", "abcdefghij", false}, // == width 10
		{"plain_overflow", "abcdefghijklmnopqrstuvwxyz", false},
		{"multi_newline", "a\nbb\n\ncccc", false},
		{"trailing_newline", "line\n", false},
		{"leading_newline", "\nline", false},
		{"dyn_simple", "[#ff0000]red[-] tail", true},
		{"dyn_wrap_midspan", "[#00ff00]aaaaaaaaaaaaaaa[-]bb", true},
		{"dyn_tag_at_boundary", "abcdefghij[#ff0000]klmno", true},
		{"dyn_only_tags", "[#ff0000][-]", true},
		{"dyn_escape", "a[[b[#ff0000]c[-]d", true},
		{"dyn_unicode", "héllo wörld with açcénts", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, width := range []int{1, 3, 10, 100} {
				tv := NewTextView().SetDynamicColors(c.dyn).SetText(c.text)
				tv.reflow(width)

				want := reference(c.text, c.dyn, width)
				if len(tv.lineSpans) != len(want) {
					t.Fatalf("width=%d line count: got %d want %d", width, len(tv.lineSpans), len(want))
				}
				for li, ls := range tv.lineSpans {
					got := tv.spanArena[ls.start:ls.end]
					wantSpans := want[li].Spans()
					if ls.width != want[li].Width() {
						t.Errorf("width=%d line %d width: got %d want %d", width, li, ls.width, want[li].Width())
					}
					if len(got) != len(wantSpans) {
						t.Fatalf("width=%d line %d span count: got %d want %d", width, li, len(got), len(wantSpans))
					}
					for si, sp := range got {
						if sp.Text != wantSpans[si].Text || sp.Style != wantSpans[si].Style {
							t.Errorf("width=%d line %d span %d: got %q/%v want %q/%v",
								width, li, si, sp.Text, sp.Style, wantSpans[si].Text, wantSpans[si].Style)
						}
					}
				}
			}
		})
	}
}

// Reusing the arenas across reflows must not corrupt output: a second reflow at
// a different width must match a fresh reference for that width.
func TestReflow_ArenaReuse(t *testing.T) {
	tv := NewTextView().SetDynamicColors(true).SetText("[#ff0000]hello[-] beautiful world")
	for _, width := range []int{40, 5, 12, 3, 100} {
		tv.reflow(width)
		want := reference(tv.text, true, width)
		if len(tv.lineSpans) != len(want) {
			t.Fatalf("width=%d: got %d lines want %d", width, len(tv.lineSpans), len(want))
		}
		for li, ls := range tv.lineSpans {
			got := tv.spanArena[ls.start:ls.end]
			ws := want[li].Spans()
			if len(got) != len(ws) {
				t.Fatalf("width=%d line %d: got %d spans want %d", width, li, len(got), len(ws))
			}
			for si := range got {
				if got[si].Text != ws[si].Text || got[si].Style != ws[si].Style {
					t.Errorf("width=%d line %d span %d mismatch", width, li, si)
				}
			}
		}
	}
}

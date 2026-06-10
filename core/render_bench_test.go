package core_test

// Benchmarks for the core render pipeline — the functions called on (nearly)
// every frame and amplified by widget-tree depth. Each render bench runs across
// three screen sizes so allocs/op and ns/op scale with cell count, not just a
// single fixed geometry.
//
// Run:  go test -run=^$ -bench=. -benchmem ./core

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

// screenSizes are the standard geometries every render bench sweeps.
var screenSizes = []struct {
	name string
	w, h int
}{
	{"80x24", 80, 24},
	{"160x48", 160, 48},
	{"240x60", 240, 60},
}

// nopScreen wraps a SimulationScreen but makes SetContent a no-op. tcell's
// SimulationScreen allocates once per SetContent, which would dominate every
// alloc count and hide dado's own allocations. A real terminal backend does
// not allocate per cell either, so suppressing it isolates the framework's
// allocations — the thing we are actually optimizing. ns/op then reflects
// dado's CPU work (layout, parsing, iteration) without backend noise.
type nopScreen struct {
	tcell.SimulationScreen
}

func (nopScreen) SetContent(_, _ int, _ rune, _ []rune, _ tcell.Style) {}

func benchScreen(w, h int) tcell.Screen {
	s := tcell.NewSimulationScreen("UTF-8")
	_ = s.Init()
	s.SetSize(w, h)
	return nopScreen{s}
}

// --- FillRect: O(w*h) cell writes, the lowest-level draw primitive ---------

func BenchmarkFillRect(b *testing.B) {
	style := tcell.StyleDefault.Background(tcell.NewRGBColor(30, 30, 30))
	for _, sz := range screenSizes {
		b.Run(sz.name, func(b *testing.B) {
			screen := benchScreen(sz.w, sz.h)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				core.FillRect(screen, 0, 0, sz.w, sz.h, ' ', style)
			}
		})
	}
}

// --- PrintTagged: parses color tags on EVERY frame. Top caching candidate. -

// taggedLine builds a width-col line peppered with color tags every `every`
// runes, exercising the tag-parse path proportional to tag density.
func taggedLine(width, every int) string {
	var sb strings.Builder
	tags := []string{"[#ff0000]", "[-]", "[::b]", "[green]"}
	for i := 0; i < width; i++ {
		if every > 0 && i%every == 0 {
			sb.WriteString(tags[(i/every)%len(tags)])
		}
		sb.WriteByte('x')
	}
	return sb.String()
}

func BenchmarkPrintTagged(b *testing.B) {
	base := tcell.StyleDefault
	cases := []struct {
		name  string
		every int // tag every N runes; 0 = no tags (plain baseline)
	}{
		{"plain", 0},
		{"sparse", 16},
		{"dense", 4},
	}
	for _, sz := range screenSizes {
		line := taggedLine(sz.w, 0)
		for _, c := range cases {
			text := line
			if c.every > 0 {
				text = taggedLine(sz.w, c.every)
			}
			b.Run(sz.name+"/"+c.name, func(b *testing.B) {
				screen := benchScreen(sz.w, sz.h)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					core.PrintTagged(screen, text, 0, 0, sz.w, base)
				}
			})
		}
	}
}

// PrintClipped is the no-parse equivalent — a floor to compare PrintTagged to.
func BenchmarkPrintClipped(b *testing.B) {
	base := tcell.StyleDefault
	for _, sz := range screenSizes {
		text := strings.Repeat("x", sz.w)
		b.Run(sz.name, func(b *testing.B) {
			screen := benchScreen(sz.w, sz.h)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				core.PrintClipped(screen, text, 0, 0, sz.w, base)
			}
		})
	}
}

// --- ParseTagged: compiles markup into a *Text once (the cached path). ------

func BenchmarkParseTagged(b *testing.B) {
	base := tcell.StyleDefault
	for _, sz := range screenSizes {
		text := taggedLine(sz.w, 8)
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = core.ParseTagged(text, base)
			}
		})
	}
}

// Text.Draw is the GOOD path: pre-parsed spans, no per-frame parsing. This
// guards against regressions that reintroduce allocation into it.
func BenchmarkText_Draw(b *testing.B) {
	base := tcell.StyleDefault
	for _, sz := range screenSizes {
		t := core.ParseTagged(taggedLine(sz.w, 8), base)
		b.Run(sz.name, func(b *testing.B) {
			screen := benchScreen(sz.w, sz.h)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t.Draw(screen, 0, 0, sz.w)
			}
		})
	}
}

// --- TextView: reflow + parse on width change, then per-frame draw ----------

func bodyText(lines, width int, dynamic bool) string {
	var sb strings.Builder
	for i := 0; i < lines; i++ {
		if dynamic {
			sb.WriteString(taggedLine(width, 12))
		} else {
			sb.WriteString(strings.Repeat("x", width))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func BenchmarkTextView_Draw(b *testing.B) {
	for _, dyn := range []bool{false, true} {
		name := "plain"
		if dyn {
			name = "dynamic"
		}
		for _, sz := range screenSizes {
			tv := core.NewTextView().SetDynamicColors(dyn).SetText(bodyText(sz.h, sz.w, dyn))
			tv.SetRect(0, 0, sz.w, sz.h)
			b.Run(name+"/"+sz.name, func(b *testing.B) {
				screen := benchScreen(sz.w, sz.h)
				tv.Draw(screen) // warm the reflow cache
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					tv.Draw(screen)
				}
			})
		}
	}
}

// Worst case: width changes every frame, forcing reflow + ParseTagged each draw.
func BenchmarkTextView_Reflow(b *testing.B) {
	for _, sz := range screenSizes {
		tv := core.NewTextView().SetDynamicColors(true).SetText(bodyText(sz.h, sz.w, true))
		b.Run(sz.name, func(b *testing.B) {
			screen := benchScreen(sz.w, sz.h)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// alternate width to invalidate the reflow cache every iteration
				w := sz.w - (i & 1)
				tv.SetRect(0, 0, w, sz.h)
				tv.Draw(screen)
			}
		})
	}
}

// --- Layout containers: Flex / Grid recompute geometry every draw ----------

func BenchmarkFlex_Draw(b *testing.B) {
	for _, n := range []int{4, 16, 64} {
		f := core.NewFlex().SetDirection(core.Column)
		for i := 0; i < n; i++ {
			f.AddItem(coretest.NewMockWidget("c"), 0, 1, false)
		}
		f.SetRect(0, 0, 160, 48)
		b.Run(itoa(n)+"children", func(b *testing.B) {
			screen := benchScreen(160, 48)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f.Draw(screen)
			}
		})
	}
}

func BenchmarkGrid_Draw(b *testing.B) {
	for _, n := range []int{4, 9, 16} {
		g := core.NewGrid()
		rows := make([]int, n)
		cols := make([]int, n)
		g.SetRows(rows...).SetColumns(cols...)
		for r := 0; r < n; r++ {
			for c := 0; c < n; c++ {
				g.AddItem(coretest.NewMockWidget("c"), r, c, 1, 1, false)
			}
		}
		g.SetRect(0, 0, 160, 48)
		b.Run(itoa(n)+"x"+itoa(n), func(b *testing.B) {
			screen := benchScreen(160, 48)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.Draw(screen)
			}
		})
	}
}

// --- core.Table: allocates width/expansion slices every draw ----------------

func BenchmarkCoreTable_Draw(b *testing.B) {
	cases := []struct{ rows, cols int }{{10, 5}, {50, 10}, {100, 20}}
	for _, c := range cases {
		t := core.NewTable()
		for r := 0; r < c.rows; r++ {
			for col := 0; col < c.cols; col++ {
				t.SetCell(r, col, core.NewTableCell("cell"))
			}
		}
		t.SetRect(0, 0, 160, 48)
		b.Run(itoa(c.rows)+"x"+itoa(c.cols), func(b *testing.B) {
			screen := benchScreen(160, 48)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t.Draw(screen)
			}
		})
	}
}

// Text.Wrap is the word-aware wrapper (distinct from the hard-wrap splitWidth
// TextView uses). Allocates styledRune + []*Text per line; benched to size it.
func BenchmarkText_Wrap(b *testing.B) {
	base := tcell.StyleDefault
	for _, sz := range screenSizes {
		// A paragraph several screens long with spaces so word-wrap has work.
		var sb strings.Builder
		for i := 0; i < sz.h*4; i++ {
			sb.WriteString("the quick brown fox jumps ")
		}
		t := core.ParseTagged(sb.String(), base)
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = t.Wrap(sz.w)
			}
		})
	}
}

// --- Full-frame integration: a realistic composite drawn end-to-end ---------

// buildFrame assembles a bordered, themed Flex(Row) of [Table | TextView] sized
// to w×h — a stand-in for a typical app frame. Returns the root, already laid
// out and warmed (one Draw) so the bench measures steady-state redraws.
func buildFrame(w, h int) core.Widget {
	bg := tcell.NewRGBColor(30, 34, 42)

	tbl := core.NewTable()
	tbl.SetBorder(true)
	tbl.SetTitle("Table")
	tbl.SetBackgroundColor(bg)
	for r := 0; r < 50; r++ {
		for c := 0; c < 4; c++ {
			tbl.SetCell(r, c, core.NewTableCell("cell "+itoa(r)+":"+itoa(c)))
		}
	}

	tv := core.NewTextView().SetDynamicColors(true).SetText(bodyText(h, w/2, true))
	tv.SetBorder(true)
	tv.SetTitle("Log")
	tv.SetBackgroundColor(bg)

	root := core.NewFlex().SetDirection(core.Row)
	root.SetBackgroundColor(bg)
	root.AddItem(tbl, 0, 1, false)
	root.AddItem(tv, 0, 1, false)
	root.SetRect(0, 0, w, h)
	return root
}

func BenchmarkApp_DrawFrame(b *testing.B) {
	for _, sz := range screenSizes {
		root := buildFrame(sz.w, sz.h)
		b.Run(sz.name, func(b *testing.B) {
			screen := benchScreen(sz.w, sz.h)
			root.Draw(screen) // warm reflow + column widths
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				root.Draw(screen)
			}
		})
	}
}

// itoa avoids fmt in benchmark labels.
func itoa(n int) string {
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

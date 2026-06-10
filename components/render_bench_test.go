package components

// Benchmarks for data-heavy scrollable component draws. Uses a nopScreen
// (SetContent no-op) so allocs/op reflects dado's own allocations, not tcell's
// per-cell SimulationScreen allocation.

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

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

// diffPair builds old/new text of n lines where every 3rd line differs, so the
// diff has a realistic mix of context/add/remove lines with line numbers.
func diffPair(n int) (string, string) {
	var oldB, newB strings.Builder
	for i := 0; i < n; i++ {
		oldB.WriteString("the quick brown fox line ")
		oldB.WriteString(itoaB(i))
		oldB.WriteByte('\n')
		if i%3 == 0 {
			newB.WriteString("the QUICK brown fox line ")
		} else {
			newB.WriteString("the quick brown fox line ")
		}
		newB.WriteString(itoaB(i))
		newB.WriteByte('\n')
	}
	return oldB.String(), newB.String()
}

func BenchmarkDiffViewer_Draw(b *testing.B) {
	for _, n := range []int{20, 200, 1000} {
		old, nw := diffPair(n)
		d := NewDiffViewer().SetShowLineNumbers(true)
		d.SetDiff(old, nw)
		d.SetRect(0, 0, 100, 40)
		b.Run(itoaB(n)+"lines", func(b *testing.B) {
			screen := benchScreen(100, 40)
			d.Draw(screen) // warm
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				d.Draw(screen)
			}
		})
	}
}

func BenchmarkDataGrid_Draw(b *testing.B) {
	cols := []GridColumn{
		{Name: "Name", Width: 16},
		{Name: "Status", Width: 10},
		{Name: "CPU", Width: 6, Align: AlignRight},
		{Name: "Mem", Width: 8, Align: AlignRight},
		{Name: "Owner", Width: 14},
		{Name: "Region", Width: 10},
	}
	for _, n := range []int{20, 200, 1000} {
		rows := make([][]GridCell, n)
		for i := 0; i < n; i++ {
			rows[i] = []GridCell{
				{Value: "service-" + itoaB(i)},
				{Value: "Running"},
				{Value: itoaB(i%100) + "%"},
				{Value: itoaB(i) + "M"},
				{Value: "team-" + itoaB(i%8)},
				{Value: "us-east"},
			}
		}
		d := NewDataGrid().SetSource(NewSliceSource(cols, rows))
		d.SetRect(0, 0, 100, 40)
		b.Run(itoaB(n)+"rows", func(b *testing.B) {
			screen := benchScreen(100, 40)
			d.Draw(screen) // warm
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				d.Draw(screen)
			}
		})
	}
}

func BenchmarkVirtualList_Draw(b *testing.B) {
	for _, n := range []int{20, 200, 1000} {
		items := make([]VirtualListItem, n)
		for i := 0; i < n; i++ {
			items[i] = VirtualListItem{ID: itoaB(i), Data: "  service-" + itoaB(i) + "          Running"}
		}
		v := NewVirtualList().SetItems(items)
		v.SetRect(0, 0, 60, 30)
		b.Run(itoaB(n)+"items", func(b *testing.B) {
			screen := benchScreen(60, 30)
			v.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v.Draw(screen)
			}
		})
	}
}

func BenchmarkLogViewer_Draw(b *testing.B) {
	levels := []LogLevel{LogLevelInfo, LogLevelDebug, LogLevelWarn, LogLevelError}
	for _, n := range []int{20, 200, 1000} {
		v := NewLogViewer()
		for i := 0; i < n; i++ {
			v.AddEntry(LogEntry{Level: levels[i%len(levels)], Message: "event " + itoaB(i) + " happened on the server", Source: "svc"})
		}
		v.SetRect(0, 0, 80, 30)
		b.Run(itoaB(n)+"entries", func(b *testing.B) {
			screen := benchScreen(80, 30)
			v.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v.Draw(screen)
			}
		})
	}
}

func BenchmarkCodeView_Draw(b *testing.B) {
	for _, n := range []int{20, 200, 1000} {
		var sb strings.Builder
		for i := 0; i < n; i++ {
			sb.WriteString("\tresult := compute(items[")
			sb.WriteString(itoaB(i))
			sb.WriteString("], cfg) // line ")
			sb.WriteString(itoaB(i))
			sb.WriteByte('\n')
		}
		c := NewCodeView().SetLanguage(LangGo).SetShowLineNumbers(true).SetCode(sb.String())
		c.SetRect(0, 0, 80, 30)
		b.Run(itoaB(n)+"lines", func(b *testing.B) {
			screen := benchScreen(80, 30)
			c.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Draw(screen)
			}
		})
	}
}

func BenchmarkTable_Draw(b *testing.B) {
	for _, n := range []int{20, 200, 1000} {
		t := NewTable().SetHeaders("Service", "Status", "Uptime", "Owner")
		for i := 0; i < n; i++ {
			t.AddRow("service-"+itoaB(i), "running", itoaB(i)+"d", "team-"+itoaB(i%8))
		}
		t.SetRect(0, 0, 80, 30)
		b.Run(itoaB(n)+"rows", func(b *testing.B) {
			screen := benchScreen(80, 30)
			t.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t.Draw(screen)
			}
		})
	}
}

func BenchmarkTree_Draw(b *testing.B) {
	for _, n := range []int{20, 200, 1000} {
		root := &TreeNode{Label: "root/", Expanded: true}
		for i := 0; i < n/10; i++ {
			dir := &TreeNode{Label: "dir-" + itoaB(i) + "/", Expanded: true}
			for j := 0; j < 10; j++ {
				dir.AddChild(&TreeNode{Label: "file-" + itoaB(j) + ".go"})
			}
			root.AddChild(dir)
		}
		tr := NewTree().SetRoot(root).SetShowLines(true).ExpandAll()
		tr.SetRect(0, 0, 60, 30)
		b.Run(itoaB(n)+"nodes", func(b *testing.B) {
			screen := benchScreen(60, 30)
			tr.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tr.Draw(screen)
			}
		})
	}
}

func BenchmarkERDGraph_Draw(b *testing.B) {
	for _, n := range []int{5, 20, 50} {
		tables := make([]*ERDTable, n)
		rels := make([]*ERDRelation, 0, n)
		for i := 0; i < n; i++ {
			tables[i] = &ERDTable{
				ID:   "t" + itoaB(i),
				Name: "table_" + itoaB(i),
				Columns: []ERDColumn{
					{Name: "id", Type: "uuid", IsPK: true},
					{Name: "name", Type: "text"},
					{Name: "owner_id", Type: "uuid", IsFK: true, FKTarget: "t0.id"},
				},
			}
			if i > 0 {
				rels = append(rels, &ERDRelation{FromTable: "t0", ToTable: "t" + itoaB(i), Cardinality: OneToMany})
			}
		}
		g := NewERDGraph().SetNodeWidth(16).SetSpacing(4, 1).SetFit(true).SetData(tables, rels)
		g.SetFocusedTable("t0")
		g.SetRect(0, 0, 120, 40)
		b.Run(itoaB(n)+"tables", func(b *testing.B) {
			screen := benchScreen(120, 40)
			g.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.Draw(screen)
			}
		})
	}
}

func BenchmarkNodeGraph_Draw(b *testing.B) {
	for _, n := range []int{5, 20, 50} {
		data := NewNodeGraphData()
		for i := 0; i < n; i++ {
			data.AddNode(&GraphNode{ID: "n" + itoaB(i), Label: "service-" + itoaB(i), Status: "running", Focused: i == 0})
		}
		for i := 1; i < n; i++ {
			data.AddEdge(&GraphEdge{From: "n0", To: "n" + itoaB(i), Label: "rpc"})
		}
		g := NewNodeGraph().SetFit(true).SetData(data)
		g.SetRect(0, 0, 120, 40)
		b.Run(itoaB(n)+"nodes", func(b *testing.B) {
			screen := benchScreen(120, 40)
			g.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.Draw(screen)
			}
		})
	}
}

func BenchmarkGraphTree_Draw(b *testing.B) {
	for _, n := range []int{5, 20, 50} {
		data := NewGraphTreeData()
		data.RootID = "n0"
		root := &GraphTreeNode{ID: "n0", Label: "root", Status: "Running", NodeType: GraphNodePrimary, Expanded: true}
		children := make([]string, 0, n)
		for i := 1; i < n; i++ {
			id := "n" + itoaB(i)
			children = append(children, id)
			data.AddNode(&GraphTreeNode{ID: id, Label: "step-" + itoaB(i), Status: "Completed", NodeType: GraphNodeSecondary, Expanded: true})
			data.AddEdge(&GraphTreeEdge{From: "n0", To: id})
		}
		root.Children = children
		data.AddNode(root)
		t := NewGraphTree().SetData(data)
		t.SetRect(0, 0, 80, 40)
		b.Run(itoaB(n)+"nodes", func(b *testing.B) {
			screen := benchScreen(80, 40)
			t.Draw(screen)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				t.Draw(screen)
			}
		})
	}
}

// itoaB is a small fmt-free int formatter for bench labels and fixtures.
func itoaB(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

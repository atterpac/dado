package inline

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableRendersHeadersAlignmentAndRows(t *testing.T) {
	t.Parallel()

	table := NewTable("Dependencies",
		TableColumn{Header: "Package"},
		TableColumn{Header: "Version", Align: AlignRight},
	).SetRows(
		[]string{"tcell", "v2.8.1"},
		[]string{"term", "v0.29.0"},
	)
	lines := plainFrameLines(table.Frame(32))
	assert.Equal(t, "Dependencies", lines[0])
	assert.Equal(t, "╭─────────┬─────────╮", lines[1])
	assert.Contains(t, lines[2], "│ Package │ Version │")
	assert.Contains(t, lines[4], "│ tcell   │  v2.8.1 │")
	assert.Equal(t, "╰─────────┴─────────╯", lines[6])
}

func TestTableShrinksColumnsAndReportsOverflow(t *testing.T) {
	t.Parallel()

	table := NewTable("Builds",
		TableColumn{Header: "Target", MinWidth: 3},
		TableColumn{Header: "Status", MinWidth: 3},
	).SetMaxRows(1).SetRows(
		[]string{"linux/amd64", "complete"},
		[]string{"darwin/arm64", "complete"},
		[]string{"windows/amd64", "failed"},
	)
	lines := plainFrameLines(table.Frame(22))
	assert.Contains(t, strings.Join(lines, "\n"), "… 2 more rows")
	assert.Contains(t, strings.Join(lines, "\n"), "…")
	for width := range 24 {
		for _, line := range plainFrameLines(table.Frame(width)) {
			assert.LessOrEqual(t, displayWidth(line), width, "width %d: %q", width, line)
		}
	}
}

func TestTableASCIITheme(t *testing.T) {
	t.Parallel()

	table := NewTable("", TableColumn{Header: "Name"}).SetTheme(ASCIIInlineTheme()).AddRow("dado")
	lines := plainFrameLines(table.Frame(12))
	assert.Equal(t, "+------+", lines[0])
	assert.True(t, strings.HasPrefix(lines[0], "+"))
	assert.True(t, strings.HasSuffix(lines[0], "+"))
	assert.Contains(t, lines[1], "| Name |")
}

func TestTableConcurrentUpdatesAndFrames(t *testing.T) {
	t.Parallel()

	table := NewTable("Concurrent", TableColumn{Header: "Value"})
	var workers sync.WaitGroup
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 50 {
				table.AddRow("value")
				_ = table.Frame(20)
			}
		}()
	}
	workers.Wait()
	assert.Len(t, plainFrameLines(table.Frame(20)), 205)
}

package components

import "sort"

// erdGridPos is a position on the layout grid.
type erdGridPos struct{ col, row int }

// computeLayout assigns x/y positions to every table node using a grid-based
// BFS algorithm that places the most-connected table at (0,0) and expands
// outward, keeping related tables adjacent.
func (g *ERDGraph) computeLayout() {
	if g.data == nil || len(g.data.tables) == 0 {
		return
	}

	g.computeNodeSizes()

	degree := make(map[string]int, len(g.data.tables))
	neighbors := make(map[string]map[string]bool, len(g.data.tables))
	for id := range g.data.tables {
		neighbors[id] = make(map[string]bool)
	}
	for _, rel := range g.data.relations {
		if _, ok := g.data.tables[rel.FromTable]; !ok {
			continue
		}
		if _, ok := g.data.tables[rel.ToTable]; !ok {
			continue
		}
		if rel.FromTable == rel.ToTable {
			continue
		}
		degree[rel.FromTable]++
		degree[rel.ToTable]++
		neighbors[rel.FromTable][rel.ToTable] = true
		neighbors[rel.ToTable][rel.FromTable] = true
	}

	sorted := make([]string, 0, len(g.data.tableOrder))
	sorted = append(sorted, g.data.tableOrder...)
	sort.Slice(sorted, func(i, j int) bool {
		return degree[sorted[i]] > degree[sorted[j]]
	})

	occupied := make(map[erdGridPos]bool)
	placement := make(map[string]erdGridPos)

	if len(sorted) > 0 {
		placement[sorted[0]] = erdGridPos{0, 0}
		occupied[erdGridPos{0, 0}] = true
	}

	queue := []string{}
	if len(sorted) > 0 {
		queue = append(queue, sorted[0])
	}
	placed := map[string]bool{}
	if len(sorted) > 0 {
		placed[sorted[0]] = true
	}

	adjacentDirs := []erdGridPos{
		{1, 0}, {0, 1}, {-1, 0}, {0, -1},
		{1, 1}, {-1, 1}, {1, -1}, {-1, -1},
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		pos := placement[current]

		for nbr := range neighbors[current] {
			if placed[nbr] {
				continue
			}
			foundSpot := false
			for _, d := range adjacentDirs {
				candidate := erdGridPos{pos.col + d.col, pos.row + d.row}
				if !occupied[candidate] {
					placement[nbr] = candidate
					occupied[candidate] = true
					placed[nbr] = true
					queue = append(queue, nbr)
					foundSpot = true
					break
				}
			}
			if !foundSpot {
				sp := erdSpiralSearch(pos, occupied)
				placement[nbr] = sp
				occupied[sp] = true
				placed[nbr] = true
				queue = append(queue, nbr)
			}
		}
	}

	for _, id := range sorted {
		if placed[id] {
			continue
		}
		sp := erdSpiralSearch(erdGridPos{0, 0}, occupied)
		placement[id] = sp
		occupied[sp] = true
		placed[id] = true
	}

	maxW, maxH := 0, 0
	for _, t := range g.data.tables {
		if t.width > maxW {
			maxW = t.width
		}
		if t.height > maxH {
			maxH = t.height
		}
	}
	cellW := maxW + g.hSpacing
	cellH := maxH + g.vSpacing

	for id, pos := range placement {
		t := g.data.tables[id]
		if t == nil {
			continue
		}
		t.x = pos.col * cellW
		t.y = pos.row * cellH
	}
}

// computeNodeSizes calculates width and height for each table based on its columns.
func (g *ERDGraph) computeNodeSizes() {
	for _, t := range g.data.tables {
		w := len([]rune(t.Name)) + 4

		for _, col := range t.Columns {
			rowW := 2 + 1 + len([]rune(col.Name)) + 2 + len([]rune(col.Type)) + 4
			if rowW > w {
				w = rowW
			}
		}

		if w < g.nodeWidth {
			w = g.nodeWidth
		}
		t.width = w

		t.height = 3 + len(t.Columns)
		if t.height < 4 {
			t.height = 4
		}
	}
}

// erdSpiralSearch finds the nearest unoccupied grid cell spiraling outward from center.
func erdSpiralSearch(center erdGridPos, occupied map[erdGridPos]bool) erdGridPos {
	for radius := 1; radius < 100; radius++ {
		for dc := -radius; dc <= radius; dc++ {
			for dr := -radius; dr <= radius; dr++ {
				if erdAbs(dc) != radius && erdAbs(dr) != radius {
					continue
				}
				candidate := erdGridPos{center.col + dc, center.row + dr}
				if !occupied[candidate] {
					return candidate
				}
			}
		}
	}
	return erdGridPos{center.col + 100, center.row}
}

func erdAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

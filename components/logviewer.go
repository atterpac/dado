package components

import (
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// LogLevel represents log severity in ascending order (Debug < Info < Warn < Error < Fatal).
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the level name
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Short returns short level name
func (l LogLevel) Short() string {
	switch l {
	case LogLevelDebug:
		return "DBG"
	case LogLevelInfo:
		return "INF"
	case LogLevelWarn:
		return "WRN"
	case LogLevelError:
		return "ERR"
	case LogLevelFatal:
		return "FTL"
	default:
		return "???"
	}
}

// LogEntry is a single log record. Fields holds structured key-value pairs
// (e.g., from a JSON log line); they are displayed alongside the message.
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Source    string            // Optional source/logger name
	Fields    map[string]string // Optional structured fields
}

// LogFilter defines filtering criteria applied to the log stream. Search and
// SearchRegex are mutually exclusive — SearchRegex takes precedence when set.
// Zero-value time fields disable time-range filtering.
type LogFilter struct {
	MinLevel    LogLevel
	MaxLevel    LogLevel
	Search      string
	SearchRegex *regexp.Regexp
	Sources     []string // Empty = all sources
	TimeFrom    time.Time
	TimeTo      time.Time
}

// LogViewer displays streaming log entries with level-based coloring, regex
// filtering, and follow mode. Append entries from any goroutine via Append;
// the viewer calls QueueUpdateDraw internally.
type LogViewer struct {
	widgetBase

	// Data
	entries    []LogEntry
	maxEntries int // Max entries to keep (0 = unlimited)

	// Filtering
	filter      LogFilter
	filteredIdx []int // Indices into entries that pass filter
	filterDirty bool

	// Display options
	showTimestamp bool
	showLevel     bool
	showSource    bool
	timestampFmt  string
	wrapLines     bool

	// Scroll state
	offsetY int
	follow  bool // Auto-scroll to bottom

	// Search
	searchPattern string
	searchMatches []int // Line indices with matches
	currentMatch  int

	// Selection
	selectedLine int

	// Callbacks
	onSelect func(entry LogEntry)
	onSearch func(pattern string, matches int)
}

// NewLogViewer creates a new log viewer component
func NewLogViewer() *LogViewer {
	v := &LogViewer{
		maxEntries:    10000,
		showTimestamp: true,
		showLevel:     true,
		timestampFmt:  "15:04:05",
		follow:        true,
		selectedLine:  -1,
		filter: LogFilter{
			MinLevel: LogLevelDebug,
			MaxLevel: LogLevelFatal,
		},
		filteredIdx: make([]int, 0),
	}
	v.initWidget()
	return v
}

// --- Configuration (Fluent API) ---

// SetMaxEntries sets the maximum log entries to keep
func (v *LogViewer) SetMaxEntries(max int) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.maxEntries = max
	return v
}

// SetShowTimestamp enables/disables timestamp display
func (v *LogViewer) SetShowTimestamp(show bool) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.showTimestamp = show
	return v
}

// SetShowLevel enables/disables level display
func (v *LogViewer) SetShowLevel(show bool) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.showLevel = show
	return v
}

// SetShowSource enables/disables source display
func (v *LogViewer) SetShowSource(show bool) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.showSource = show
	return v
}

// SetTimestampFormat sets the timestamp format
func (v *LogViewer) SetTimestampFormat(format string) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.timestampFmt = format
	return v
}

// SetWrapLines enables/disables line wrapping
func (v *LogViewer) SetWrapLines(wrap bool) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.wrapLines = wrap
	return v
}

// SetFollow enables/disables auto-scroll
func (v *LogViewer) SetFollow(follow bool) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.follow = follow
	return v
}

// SetOnSelect sets callback for entry selection
func (v *LogViewer) SetOnSelect(fn func(entry LogEntry)) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.onSelect = fn
	return v
}

// SetOnSearch sets callback for search results
func (v *LogViewer) SetOnSearch(fn func(pattern string, matches int)) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.onSearch = fn
	return v
}

// --- Filtering ---

// SetFilter sets the log filter
func (v *LogViewer) SetFilter(filter LogFilter) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.filter = filter
	v.filterDirty = true
	return v
}

// SetMinLevel sets minimum log level to show
func (v *LogViewer) SetMinLevel(level LogLevel) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.filter.MinLevel = level
	v.filterDirty = true
	return v
}

// SetSearch sets the search pattern
func (v *LogViewer) SetSearch(pattern string) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.filter.Search = pattern
	if pattern != "" {
		v.filter.SearchRegex, _ = regexp.Compile("(?i)" + regexp.QuoteMeta(pattern))
	} else {
		v.filter.SearchRegex = nil
	}
	v.filterDirty = true
	return v
}

// SetSearchRegex sets a regex search pattern
func (v *LogViewer) SetSearchRegex(pattern string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	v.filter.Search = pattern
	v.filter.SearchRegex = re
	v.filterDirty = true
	return nil
}

// ClearFilter removes all filters
func (v *LogViewer) ClearFilter() *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.filter = LogFilter{
		MinLevel: LogLevelDebug,
		MaxLevel: LogLevelFatal,
	}
	v.filterDirty = true
	return v
}

// --- Data Management ---

// AddEntry adds a log entry
func (v *LogViewer) AddEntry(entry LogEntry) *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.entries = append(v.entries, entry)

	// Trim if over limit
	if v.maxEntries > 0 && len(v.entries) > v.maxEntries {
		excess := len(v.entries) - v.maxEntries
		newEntries := make([]LogEntry, v.maxEntries)
		copy(newEntries, v.entries[excess:])
		v.entries = newEntries
		// Adjust filtered indices
		newFiltered := make([]int, 0, len(v.filteredIdx))
		for _, idx := range v.filteredIdx {
			if idx >= excess {
				newFiltered = append(newFiltered, idx-excess)
			}
		}
		v.filteredIdx = newFiltered
	}

	// Check if new entry passes filter
	if v.passesFilter(entry) {
		v.filteredIdx = append(v.filteredIdx, len(v.entries)-1)
	}

	// Auto-scroll if following
	if v.follow {
		_, _, _, height := v.GetInnerRect()
		if len(v.filteredIdx) > height {
			v.offsetY = len(v.filteredIdx) - height
		}
	}

	return v
}

// Log adds a simple log entry
func (v *LogViewer) Log(level LogLevel, message string) *LogViewer {
	return v.AddEntry(LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	})
}

// Debug logs a debug message
func (v *LogViewer) Debug(message string) *LogViewer {
	return v.Log(LogLevelDebug, message)
}

// Info logs an info message
func (v *LogViewer) Info(message string) *LogViewer {
	return v.Log(LogLevelInfo, message)
}

// Warn logs a warning message
func (v *LogViewer) Warn(message string) *LogViewer {
	return v.Log(LogLevelWarn, message)
}

// Error logs an error message
func (v *LogViewer) Error(message string) *LogViewer {
	return v.Log(LogLevelError, message)
}

// Clear removes all entries
func (v *LogViewer) Clear() *LogViewer {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.entries = nil
	v.filteredIdx = nil
	v.offsetY = 0
	return v
}

// GetEntries returns all entries
func (v *LogViewer) GetEntries() []LogEntry {
	v.mu.RLock()
	defer v.mu.RUnlock()
	result := make([]LogEntry, len(v.entries))
	copy(result, v.entries)
	return result
}

// GetFilteredEntries returns filtered entries
func (v *LogViewer) GetFilteredEntries() []LogEntry {
	v.mu.RLock()
	defer v.mu.RUnlock()
	v.applyFilter()
	result := make([]LogEntry, len(v.filteredIdx))
	for i, idx := range v.filteredIdx {
		result[i] = v.entries[idx]
	}
	return result
}

// EntryCount returns total entry count
func (v *LogViewer) EntryCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.entries)
}

// FilteredCount returns filtered entry count
func (v *LogViewer) FilteredCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	v.applyFilter()
	return len(v.filteredIdx)
}

// --- Internal ---

func (v *LogViewer) passesFilter(entry LogEntry) bool {
	// Level filter
	if entry.Level < v.filter.MinLevel || entry.Level > v.filter.MaxLevel {
		return false
	}

	// Source filter
	if len(v.filter.Sources) > 0 {
		found := false
		for _, s := range v.filter.Sources {
			if entry.Source == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Time filter
	if !v.filter.TimeFrom.IsZero() && entry.Timestamp.Before(v.filter.TimeFrom) {
		return false
	}
	if !v.filter.TimeTo.IsZero() && entry.Timestamp.After(v.filter.TimeTo) {
		return false
	}

	// Search filter
	if v.filter.SearchRegex != nil {
		if !v.filter.SearchRegex.MatchString(entry.Message) {
			return false
		}
	} else if v.filter.Search != "" {
		if !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(v.filter.Search)) {
			return false
		}
	}

	return true
}

func (v *LogViewer) applyFilter() {
	if !v.filterDirty {
		return
	}

	v.filteredIdx = make([]int, 0, len(v.entries))
	for i, entry := range v.entries {
		if v.passesFilter(entry) {
			v.filteredIdx = append(v.filteredIdx, i)
		}
	}
	v.filterDirty = false

	// Adjust offset if needed
	_, _, _, height := v.GetInnerRect()
	if height > 0 && v.offsetY > len(v.filteredIdx)-height {
		v.offsetY = len(v.filteredIdx) - height
		if v.offsetY < 0 {
			v.offsetY = 0
		}
	}
}

func (v *LogViewer) getLevelColor(level LogLevel) tcell.Color {
	th := v.th()
	switch level {
	case LogLevelDebug:
		return th.FgDim()
	case LogLevelInfo:
		return th.Info()
	case LogLevelWarn:
		return th.Warning()
	case LogLevelError, LogLevelFatal:
		return th.Error()
	default:
		return th.Fg()
	}
}

// Draw renders the log viewer
func (v *LogViewer) Draw(screen tcell.Screen) {
	v.Box.DrawForSubclass(screen)
	x, y, width, height := v.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	v.mu.Lock()
	v.applyFilter()
	v.mu.Unlock()

	v.mu.RLock()
	defer v.mu.RUnlock()

	th := v.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	highlightBg := th.BgLight()

	bgStyle := tcell.StyleDefault.Background(bgColor)

	// Clear area
	fillRect(screen, x, y, width, height, bgStyle)

	if len(v.filteredIdx) == 0 {
		// Show empty state
		msg := "No log entries"
		if v.filter.Search != "" {
			msg = "No matching entries"
		}
		msgStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		msgX := x + (width-len(msg))/2
		for i, r := range msg {
			if msgX+i < x+width {
				screen.SetContent(msgX+i, y+height/2, r, nil, msgStyle)
			}
		}
		return
	}

	// Draw visible entries. The viewport translates content (column, rowIdx)
	// to screen coordinates and clips writes past the right/bottom edges.
	vp := core.NewViewport(x, y, width, height)
	vp.SetContentSize(width, len(v.filteredIdx))
	vp.SetOffset(0, v.offsetY)
	_, v.offsetY = vp.Offset() // keep field in sync

	first, last := vp.VisibleRows()
	for rowIdx := first; rowIdx < last; rowIdx++ {
		entryIdx := v.filteredIdx[rowIdx]
		entry := v.entries[entryIdx]

		isSelected := rowIdx == v.selectedLine
		rowBg := bgColor
		if isSelected {
			rowBg = highlightBg
		}

		col := 0

		// Timestamp
		if v.showTimestamp {
			ts := entry.Timestamp.Format(v.timestampFmt)
			tsStyle := tcell.StyleDefault.Background(rowBg).Foreground(fgDimColor)
			for _, r := range ts {
				vp.SetContent(screen, col, rowIdx, r, tsStyle)
				col++
			}
			col++ // space
		}

		// Level
		if v.showLevel {
			levelColor := v.getLevelColor(entry.Level)
			levelStyle := tcell.StyleDefault.Background(rowBg).Foreground(levelColor)
			levelStr := entry.Level.Short()
			for _, r := range levelStr {
				vp.SetContent(screen, col, rowIdx, r, levelStyle)
				col++
			}
			col++ // space
		}

		// Source
		if v.showSource && entry.Source != "" {
			srcStyle := tcell.StyleDefault.Background(rowBg).Foreground(fgDimColor)
			srcStr := "[" + entry.Source + "]"
			for _, r := range srcStr {
				vp.SetContent(screen, col, rowIdx, r, srcStyle)
				col++
			}
			col++ // space
		}

		// Message
		msgStyle := tcell.StyleDefault.Background(rowBg).Foreground(fgColor)
		for _, r := range entry.Message {
			vp.SetContent(screen, col, rowIdx, r, msgStyle)
			col++
		}

		// Fill rest of selected line
		if isSelected {
			fillStyle := tcell.StyleDefault.Background(rowBg)
			for ; col < width; col++ {
				vp.SetContent(screen, col, rowIdx, ' ', fillStyle)
			}
		}
	}

	// Draw scrollbar if needed
	if len(v.filteredIdx) > height {
		v.drawScrollbar(screen, x+width-1, y, height)
	}
}

func (v *LogViewer) drawScrollbar(screen tcell.Screen, x, y, height int) {
	th := v.th()
	trackColor := th.BgLight()
	thumbColor := th.FgDim()

	thumbSize := height * height / len(v.filteredIdx)
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > height {
		thumbSize = height
	}

	thumbPos := 0
	if len(v.filteredIdx) > height {
		thumbPos = v.offsetY * (height - thumbSize) / (len(v.filteredIdx) - height)
	}

	trackStyle := tcell.StyleDefault.Background(trackColor)
	thumbStyle := tcell.StyleDefault.Background(thumbColor)

	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			screen.SetContent(x, y+i, ' ', nil, thumbStyle)
		} else {
			screen.SetContent(x, y+i, ' ', nil, trackStyle)
		}
	}
}

func (v *LogViewer) HandleKey(ev *tcell.EventKey) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	_, _, _, height := v.GetInnerRect()

	switch ev.Key() {
	case tcell.KeyDown:
		if v.offsetY < len(v.filteredIdx)-1 {
			v.offsetY++
			v.follow = false
		}
		return true
	case tcell.KeyUp:
		if v.offsetY > 0 {
			v.offsetY--
			v.follow = false
		}
		return true
	case tcell.KeyPgDn:
		v.offsetY += height
		if v.offsetY > len(v.filteredIdx)-height {
			v.offsetY = len(v.filteredIdx) - height
		}
		if v.offsetY < 0 {
			v.offsetY = 0
		}
		v.follow = false
		return true
	case tcell.KeyPgUp:
		v.offsetY -= height
		if v.offsetY < 0 {
			v.offsetY = 0
		}
		v.follow = false
		return true
	case tcell.KeyHome:
		v.offsetY = 0
		v.follow = false
		return true
	case tcell.KeyEnd:
		v.offsetY = len(v.filteredIdx) - height
		if v.offsetY < 0 {
			v.offsetY = 0
		}
		v.follow = true
		return true
	case tcell.KeyEnter:
		if v.selectedLine >= 0 && v.selectedLine < len(v.filteredIdx) {
			entry := v.entries[v.filteredIdx[v.selectedLine]]
			if v.onSelect != nil {
				v.onSelect(entry)
			}
		}
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			if v.offsetY < len(v.filteredIdx)-1 {
				v.offsetY++
				v.follow = false
			}
			return true
		case 'k':
			if v.offsetY > 0 {
				v.offsetY--
				v.follow = false
			}
			return true
		case 'g':
			v.offsetY = 0
			v.follow = false
			return true
		case 'G':
			v.offsetY = len(v.filteredIdx) - height
			if v.offsetY < 0 {
				v.offsetY = 0
			}
			v.follow = true
			return true
		case 'f':
			v.follow = !v.follow
			return true
		case 'c':
			v.entries = nil
			v.filteredIdx = nil
			v.offsetY = 0
			return true
		}
	}
	return false
}

// GetFieldHeight returns preferred height
func (v *LogViewer) GetFieldHeight() int {
	return 15
}

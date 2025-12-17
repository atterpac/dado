package recipes

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/input"
	"github.com/atterpac/jig/nav"
	"github.com/atterpac/jig/theme"
)

// LogLine represents a single log entry.
type LogLine struct {
	Timestamp string
	Level     string
	Message   string
	Raw       string
}

// LogViewer is a streaming log display with rich features.
type LogViewer struct {
	*tview.Box

	// Data
	lines    []LogLine
	maxLines int
	offset   int // scroll offset
	follow   bool
	paused   bool

	// Search
	searchQuery   string
	searchResults []int
	searchIndex   int

	// Display options
	showTimestamps bool
	wrap           bool
	levelColors    map[string]tcell.Color

	// Components
	searchBar *input.CommandBar
	hints     []components.KeyHint

	// Callbacks
	onSearch func(query string, lineNum int)

	mu sync.RWMutex
}

// NewLogViewer creates a new LogViewer.
func NewLogViewer() *LogViewer {
	l := &LogViewer{
		Box:            tview.NewBox(),
		maxLines:       10000,
		follow:         true,
		showTimestamps: true,
		levelColors: map[string]tcell.Color{
			"ERROR": theme.Error(),
			"WARN":  theme.Warning(),
			"INFO":  theme.Info(),
			"DEBUG": theme.FgDim(),
		},
		searchBar: input.NewCommandBar(),
	}

	// Setup search bar
	l.searchBar.Hide()
	l.searchBar.SetOnSubmit(func(cmdType input.CommandType, cmd string) {
		l.searchQuery = cmd
		l.performSearch()
		l.searchBar.Hide()
	})
	l.searchBar.SetOnCancel(func() {
		l.searchBar.Hide()
	})
	l.searchBar.SetOnChange(func(cmd string) {
		// Live search
		l.searchQuery = cmd
		l.performSearch()
	})

	// Default hints
	l.updateHints()

	return l
}

// SetMaxLines sets the maximum number of lines to keep.
func (l *LogViewer) SetMaxLines(max int) *LogViewer {
	l.maxLines = max
	return l
}

// SetFollow enables/disables auto-follow.
func (l *LogViewer) SetFollow(follow bool) *LogViewer {
	l.follow = follow
	return l
}

// SetShowTimestamps enables/disables timestamp display.
func (l *LogViewer) SetShowTimestamps(show bool) *LogViewer {
	l.showTimestamps = show
	return l
}

// SetWrap enables/disables line wrapping.
func (l *LogViewer) SetWrap(wrap bool) *LogViewer {
	l.wrap = wrap
	return l
}

// SetLevelColors sets the log level colors.
func (l *LogViewer) SetLevelColors(colors map[string]tcell.Color) *LogViewer {
	l.levelColors = colors
	return l
}

// SetOnSearch sets the callback for search matches.
func (l *LogViewer) SetOnSearch(fn func(query string, lineNum int)) *LogViewer {
	l.onSearch = fn
	return l
}

// Append adds a new log line.
func (l *LogViewer) Append(line string) *LogViewer {
	l.mu.Lock()
	defer l.mu.Unlock()

	logLine := l.parseLine(line)
	l.lines = append(l.lines, logLine)

	// Trim to max lines
	if len(l.lines) > l.maxLines {
		l.lines = l.lines[len(l.lines)-l.maxLines:]
	}

	// Auto-scroll if following
	if l.follow && !l.paused {
		_, _, _, height := l.GetInnerRect()
		l.offset = len(l.lines) - height
		if l.offset < 0 {
			l.offset = 0
		}
	}

	return l
}

// AppendLines adds multiple log lines.
func (l *LogViewer) AppendLines(lines []string) *LogViewer {
	for _, line := range lines {
		l.Append(line)
	}
	return l
}

// Clear removes all log lines.
func (l *LogViewer) Clear() *LogViewer {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = nil
	l.offset = 0
	l.searchResults = nil
	return l
}

// Pause pauses auto-follow.
func (l *LogViewer) Pause() *LogViewer {
	l.paused = true
	return l
}

// Resume resumes auto-follow.
func (l *LogViewer) Resume() *LogViewer {
	l.paused = false
	if l.follow {
		l.scrollToBottom()
	}
	return l
}

// IsPaused returns whether auto-follow is paused.
func (l *LogViewer) IsPaused() bool {
	return l.paused
}

func (l *LogViewer) parseLine(line string) LogLine {
	logLine := LogLine{Raw: line}

	parts := strings.Fields(line)
	if len(parts) >= 2 {
		// Check if first part looks like a timestamp
		if len(parts[0]) >= 8 && (strings.Contains(parts[0], ":") || strings.Contains(parts[0], "-")) {
			logLine.Timestamp = parts[0]
			parts = parts[1:]
		}

		// Check for level
		if len(parts) > 0 {
			level := strings.ToUpper(strings.Trim(parts[0], "[]():"))
			if level == "ERROR" || level == "WARN" || level == "WARNING" || level == "INFO" || level == "DEBUG" || level == "TRACE" {
				logLine.Level = level
				if level == "WARNING" {
					logLine.Level = "WARN"
				}
				parts = parts[1:]
			}
		}

		logLine.Message = strings.Join(parts, " ")
	} else {
		logLine.Message = line
	}

	return logLine
}

func (l *LogViewer) performSearch() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.searchResults = nil
	l.searchIndex = 0

	if l.searchQuery == "" {
		return
	}

	query := strings.ToLower(l.searchQuery)
	for i, line := range l.lines {
		if strings.Contains(strings.ToLower(line.Raw), query) {
			l.searchResults = append(l.searchResults, i)
		}
	}

	// Jump to first result
	if len(l.searchResults) > 0 {
		l.jumpToSearchResult(0)
	}
}

func (l *LogViewer) jumpToSearchResult(index int) {
	if index < 0 || index >= len(l.searchResults) {
		return
	}
	l.searchIndex = index
	lineNum := l.searchResults[index]
	l.offset = lineNum

	// Center the result
	_, _, _, height := l.GetInnerRect()
	l.offset = lineNum - height/2
	if l.offset < 0 {
		l.offset = 0
	}

	if l.onSearch != nil {
		l.onSearch(l.searchQuery, lineNum)
	}
}

func (l *LogViewer) nextSearchResult() {
	if len(l.searchResults) == 0 {
		return
	}
	l.jumpToSearchResult((l.searchIndex + 1) % len(l.searchResults))
}

func (l *LogViewer) prevSearchResult() {
	if len(l.searchResults) == 0 {
		return
	}
	idx := l.searchIndex - 1
	if idx < 0 {
		idx = len(l.searchResults) - 1
	}
	l.jumpToSearchResult(idx)
}

func (l *LogViewer) scrollToBottom() {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, _, _, height := l.GetInnerRect()
	l.offset = len(l.lines) - height
	if l.offset < 0 {
		l.offset = 0
	}
}

func (l *LogViewer) updateHints() {
	l.hints = []components.KeyHint{
		{Key: "/", Description: "Search"},
		{Key: "n/N", Description: "Next/Prev match"},
		{Key: "f", Description: "Toggle follow"},
		{Key: "p", Description: "Pause/Resume"},
		{Key: "t", Description: "Toggle timestamps"},
		{Key: "w", Description: "Toggle wrap"},
		{Key: "c", Description: "Clear"},
		{Key: "g/G", Description: "Top/Bottom"},
	}
}

// Draw renders the log viewer.
func (l *LogViewer) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, width, height := l.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	// Get colors
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	highlightColor := theme.Highlight()

	// Reserve space for search bar and hints
	logHeight := height - 2
	searchBarY := y
	hintsY := y + height - 1

	// Draw search bar
	l.searchBar.SetRect(x, searchBarY, width, 1)
	l.searchBar.Draw(screen)

	// Draw hints
	hintBar := components.NewKeyHintBar()
	hintBar.SetHints(l.hints)
	hintBar.SetRect(x, hintsY, width, 1)
	hintBar.Draw(screen)

	// Draw status indicator
	statusStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)
	if l.paused {
		statusStyle = tcell.StyleDefault.Background(bgColor).Foreground(theme.Warning())
		screen.SetContent(x+width-3, searchBarY, '⏸', nil, statusStyle)
	} else if l.follow {
		screen.SetContent(x+width-3, searchBarY, '▼', nil, statusStyle)
	}

	// Draw search result count
	if len(l.searchResults) > 0 {
		countStr := logItoa(l.searchIndex+1) + "/" + logItoa(len(l.searchResults))
		countStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		col := x + width - len(countStr) - 5
		for _, r := range countStr {
			screen.SetContent(col, searchBarY, r, nil, countStyle)
			col++
		}
	}

	// Ensure offset is valid
	maxOffset := len(l.lines) - logHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if l.offset > maxOffset {
		l.offset = maxOffset
	}
	if l.offset < 0 {
		l.offset = 0
	}

	// Draw log lines
	logY := y + 1
	for i := 0; i < logHeight && l.offset+i < len(l.lines); i++ {
		line := l.lines[l.offset+i]
		row := logY + i

		// Clear row
		clearStyle := tcell.StyleDefault.Background(bgColor)
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, clearStyle)
		}

		col := x

		// Draw timestamp
		if l.showTimestamps && line.Timestamp != "" {
			timestampStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			for _, r := range line.Timestamp {
				if col < x+width {
					screen.SetContent(col, row, r, nil, timestampStyle)
					col++
				}
			}
			screen.SetContent(col, row, ' ', nil, clearStyle)
			col++
		}

		// Draw level
		if line.Level != "" {
			levelColor := fgColor
			if c, ok := l.levelColors[line.Level]; ok {
				levelColor = c
			}
			levelStyle := tcell.StyleDefault.Background(bgColor).Foreground(levelColor)
			for _, r := range "[" + line.Level + "]" {
				if col < x+width {
					screen.SetContent(col, row, r, nil, levelStyle)
					col++
				}
			}
			screen.SetContent(col, row, ' ', nil, clearStyle)
			col++
		}

		// Draw message with search highlighting
		messageStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		highlightStyle := tcell.StyleDefault.Background(highlightColor).Foreground(bgColor)

		message := line.Message
		if line.Message == "" {
			message = line.Raw
		}

		// Check if this line is a search result
		isSearchResult := false
		for _, resultIdx := range l.searchResults {
			if resultIdx == l.offset+i {
				isSearchResult = true
				break
			}
		}

		if isSearchResult && l.searchQuery != "" {
			// Highlight matches
			messageLower := strings.ToLower(message)
			queryLower := strings.ToLower(l.searchQuery)
			lastEnd := 0

			for {
				idx := strings.Index(messageLower[lastEnd:], queryLower)
				if idx == -1 {
					// Draw remaining text
					for _, r := range message[lastEnd:] {
						if col < x+width {
							screen.SetContent(col, row, r, nil, messageStyle)
							col++
						}
					}
					break
				}

				matchStart := lastEnd + idx
				matchEnd := matchStart + len(l.searchQuery)

				// Draw text before match
				for _, r := range message[lastEnd:matchStart] {
					if col < x+width {
						screen.SetContent(col, row, r, nil, messageStyle)
						col++
					}
				}

				// Draw match highlighted
				for _, r := range message[matchStart:matchEnd] {
					if col < x+width {
						screen.SetContent(col, row, r, nil, highlightStyle)
						col++
					}
				}

				lastEnd = matchEnd
			}
		} else {
			for _, r := range message {
				if col < x+width {
					screen.SetContent(col, row, r, nil, messageStyle)
					col++
				}
			}
		}
	}
}

// InputHandler handles keyboard input.
func (l *LogViewer) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// Handle search bar if visible
		if l.searchBar.IsVisible() {
			if handler := l.searchBar.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
			return
		}

		_, _, _, height := l.GetInnerRect()
		logHeight := height - 2

		switch event.Key() {
		case tcell.KeyUp:
			l.offset--
			l.paused = true
		case tcell.KeyDown:
			l.offset++
		case tcell.KeyPgUp:
			l.offset -= logHeight
			l.paused = true
		case tcell.KeyPgDn:
			l.offset += logHeight
		case tcell.KeyHome:
			l.offset = 0
			l.paused = true
		case tcell.KeyEnd:
			l.scrollToBottom()
			l.paused = false
		case tcell.KeyRune:
			switch event.Rune() {
			case '/':
				l.searchBar.Show(input.CommandTypeSearch)
			case 'n':
				l.nextSearchResult()
			case 'N':
				l.prevSearchResult()
			case 'f':
				l.follow = !l.follow
				if l.follow {
					l.scrollToBottom()
				}
			case 'p':
				if l.paused {
					l.Resume()
				} else {
					l.Pause()
				}
			case 't':
				l.showTimestamps = !l.showTimestamps
			case 'w':
				l.wrap = !l.wrap
			case 'c':
				l.Clear()
			case 'g':
				l.offset = 0
				l.paused = true
			case 'G':
				l.scrollToBottom()
				l.paused = false
			case 'j':
				l.offset++
			case 'k':
				l.offset--
				l.paused = true
			}
		case tcell.KeyCtrlD:
			l.offset += logHeight / 2
		case tcell.KeyCtrlU:
			l.offset -= logHeight / 2
			l.paused = true
		case tcell.KeyEscape:
			if l.searchQuery != "" {
				l.searchQuery = ""
				l.searchResults = nil
			}
		}
	})
}

// Start begins the log viewer lifecycle.
func (l *LogViewer) Start() {}

// Stop ends the log viewer lifecycle.
func (l *LogViewer) Stop() {}

// Hints returns the current key hints.
func (l *LogViewer) Hints() []components.KeyHint {
	return l.hints
}

// Focus handles focus.
func (l *LogViewer) Focus(delegate func(tview.Primitive)) {
	if l.searchBar.IsVisible() {
		delegate(l.searchBar)
	} else {
		l.Box.Focus(delegate)
	}
}

// HasFocus returns whether the viewer has focus.
func (l *LogViewer) HasFocus() bool {
	return l.Box.HasFocus() || l.searchBar.HasFocus()
}

// logItoa converts int to string.
func logItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// Ensure LogViewer implements nav.Component
var _ nav.Component = (*LogViewer)(nil)

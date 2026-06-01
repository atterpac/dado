package components

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// Language defines supported syntax highlighting languages
type Language string

const (
	LangNone       Language = ""
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
	LangRust       Language = "rust"
	LangJSON       Language = "json"
	LangYAML       Language = "yaml"
	LangSQL        Language = "sql"
	LangBash       Language = "bash"
	LangMarkdown   Language = "markdown"
)

// TokenType represents syntax token categories
type TokenType int

const (
	TokenNormal TokenType = iota
	TokenKeyword
	TokenString
	TokenNumber
	TokenComment
	TokenFunction
	TokenType_
	TokenOperator
	TokenPunctuation
	TokenBuiltin
)

// Token represents a syntax-highlighted token
type Token struct {
	Text  string
	Type  TokenType
	Start int
	End   int
}

// CodeView displays syntax-highlighted code
type CodeView struct {
	widgetBase

	// Content
	lines    []string
	language Language

	// Display options
	showLineNumbers bool
	lineNumberWidth int
	tabWidth        int
	wrapLines       bool
	highlightLine   int // -1 = none

	// Scroll state
	offsetX int
	offsetY int

	// Selection
	selectionStart [2]int // [line, col]
	selectionEnd   [2]int
	hasSelection   bool

	// Cached tokens per line
	tokenCache map[int][]Token

	// Callbacks
	onLineClick func(line int)
}

// NewCodeView creates a new code view component
func NewCodeView() *CodeView {
	c := &CodeView{
		showLineNumbers: true,
		tabWidth:        4,
		highlightLine:   -1,
		tokenCache:      make(map[int][]Token),
	}
	c.initWidget()
	return c
}

// --- Configuration (Fluent API) ---

// SetCode sets the code content
func (c *CodeView) SetCode(code string) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lines = strings.Split(code, "\n")
	c.tokenCache = make(map[int][]Token) // Clear cache
	c.offsetX = 0
	c.offsetY = 0
	return c
}

// SetLines sets code from line slice
func (c *CodeView) SetLines(lines []string) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lines = lines
	c.tokenCache = make(map[int][]Token)
	return c
}

// SetLanguage sets the syntax highlighting language
func (c *CodeView) SetLanguage(lang Language) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.language = lang
	c.tokenCache = make(map[int][]Token) // Re-tokenize
	return c
}

// SetShowLineNumbers enables/disables line numbers
func (c *CodeView) SetShowLineNumbers(show bool) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.showLineNumbers = show
	return c
}

// SetTabWidth sets the tab display width
func (c *CodeView) SetTabWidth(width int) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tabWidth = width
	return c
}

// SetWrapLines enables/disables line wrapping
func (c *CodeView) SetWrapLines(wrap bool) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wrapLines = wrap
	return c
}

// SetHighlightLine sets the highlighted line (1-based, -1 to disable)
func (c *CodeView) SetHighlightLine(line int) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.highlightLine = line
	return c
}

// SetOnLineClick sets callback for line clicks
func (c *CodeView) SetOnLineClick(fn func(line int)) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onLineClick = fn
	return c
}

// ScrollTo scrolls to a specific line
func (c *CodeView) ScrollTo(line int) *CodeView {
	c.mu.Lock()
	defer c.mu.Unlock()
	if line < 0 {
		line = 0
	}
	if line >= len(c.lines) {
		line = len(c.lines) - 1
	}
	c.offsetY = line
	return c
}

// --- Data Access ---

// GetLines returns the code lines
func (c *CodeView) GetLines() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.lines))
	copy(result, c.lines)
	return result
}

// GetLine returns a specific line (0-based)
func (c *CodeView) GetLine(index int) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if index >= 0 && index < len(c.lines) {
		return c.lines[index]
	}
	return ""
}

// LineCount returns the number of lines
func (c *CodeView) LineCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.lines)
}

// --- Tokenization ---

func (c *CodeView) tokenizeLine(lineIdx int, line string) []Token {
	// Check cache
	if tokens, ok := c.tokenCache[lineIdx]; ok {
		return tokens
	}

	var tokens []Token

	switch c.language {
	case LangGo:
		tokens = c.tokenizeGo(line)
	case LangPython:
		tokens = c.tokenizePython(line)
	case LangJavaScript, LangTypeScript:
		tokens = c.tokenizeJS(line)
	case LangJSON:
		tokens = c.tokenizeJSON(line)
	case LangBash:
		tokens = c.tokenizeBash(line)
	default:
		// No highlighting
		tokens = []Token{{Text: line, Type: TokenNormal, Start: 0, End: len(line)}}
	}

	c.tokenCache[lineIdx] = tokens
	return tokens
}

func (c *CodeView) tokenizeGo(line string) []Token {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}

	builtins := map[string]bool{
		"true": true, "false": true, "nil": true, "iota": true,
		"append": true, "cap": true, "close": true, "copy": true, "delete": true,
		"len": true, "make": true, "new": true, "panic": true, "recover": true,
		"print": true, "println": true,
	}

	types := map[string]bool{
		"bool": true, "byte": true, "complex64": true, "complex128": true,
		"error": true, "float32": true, "float64": true, "int": true,
		"int8": true, "int16": true, "int32": true, "int64": true,
		"rune": true, "string": true, "uint": true, "uint8": true,
		"uint16": true, "uint32": true, "uint64": true, "uintptr": true,
	}

	return c.tokenizeGeneric(line, keywords, builtins, types, "//", []string{"/*", "*/"}, '"', '\'', '`')
}

func (c *CodeView) tokenizePython(line string) []Token {
	keywords := map[string]bool{
		"and": true, "as": true, "assert": true, "async": true, "await": true,
		"break": true, "class": true, "continue": true, "def": true, "del": true,
		"elif": true, "else": true, "except": true, "finally": true, "for": true,
		"from": true, "global": true, "if": true, "import": true, "in": true,
		"is": true, "lambda": true, "nonlocal": true, "not": true, "or": true,
		"pass": true, "raise": true, "return": true, "try": true, "while": true,
		"with": true, "yield": true,
	}

	builtins := map[string]bool{
		"True": true, "False": true, "None": true,
		"print": true, "len": true, "range": true, "type": true, "str": true,
		"int": true, "float": true, "list": true, "dict": true, "set": true,
		"tuple": true, "bool": true, "open": true, "input": true,
	}

	return c.tokenizeGeneric(line, keywords, builtins, nil, "#", nil, '"', '\'', 0)
}

func (c *CodeView) tokenizeJS(line string) []Token {
	keywords := map[string]bool{
		"break": true, "case": true, "catch": true, "class": true, "const": true,
		"continue": true, "debugger": true, "default": true, "delete": true, "do": true,
		"else": true, "export": true, "extends": true, "finally": true, "for": true,
		"function": true, "if": true, "import": true, "in": true, "instanceof": true,
		"let": true, "new": true, "return": true, "super": true, "switch": true,
		"this": true, "throw": true, "try": true, "typeof": true, "var": true,
		"void": true, "while": true, "with": true, "yield": true, "async": true,
		"await": true,
	}

	builtins := map[string]bool{
		"true": true, "false": true, "null": true, "undefined": true,
		"console": true, "window": true, "document": true, "JSON": true,
		"Array": true, "Object": true, "String": true, "Number": true,
		"Boolean": true, "Promise": true, "Map": true, "Set": true,
	}

	return c.tokenizeGeneric(line, keywords, builtins, nil, "//", []string{"/*", "*/"}, '"', '\'', '`')
}

func (c *CodeView) tokenizeJSON(line string) []Token {
	var tokens []Token
	i := 0
	n := len(line)

	for i < n {
		ch := line[i]

		// Whitespace
		if ch == ' ' || ch == '\t' {
			start := i
			for i < n && (line[i] == ' ' || line[i] == '\t') {
				i++
			}
			tokens = append(tokens, Token{Text: line[start:i], Type: TokenNormal, Start: start, End: i})
			continue
		}

		// String
		if ch == '"' {
			start := i
			i++
			for i < n && line[i] != '"' {
				if line[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			if i < n {
				i++ // closing quote
			}
			tokens = append(tokens, Token{Text: line[start:i], Type: TokenString, Start: start, End: i})
			continue
		}

		// Number
		if (ch >= '0' && ch <= '9') || ch == '-' {
			start := i
			if ch == '-' {
				i++
			}
			for i < n && ((line[i] >= '0' && line[i] <= '9') || line[i] == '.' || line[i] == 'e' || line[i] == 'E' || line[i] == '+' || line[i] == '-') {
				i++
			}
			tokens = append(tokens, Token{Text: line[start:i], Type: TokenNumber, Start: start, End: i})
			continue
		}

		// Keywords: true, false, null
		if i+4 <= n && line[i:i+4] == "true" {
			tokens = append(tokens, Token{Text: "true", Type: TokenBuiltin, Start: i, End: i + 4})
			i += 4
			continue
		}
		if i+5 <= n && line[i:i+5] == "false" {
			tokens = append(tokens, Token{Text: "false", Type: TokenBuiltin, Start: i, End: i + 5})
			i += 5
			continue
		}
		if i+4 <= n && line[i:i+4] == "null" {
			tokens = append(tokens, Token{Text: "null", Type: TokenBuiltin, Start: i, End: i + 4})
			i += 4
			continue
		}

		// Punctuation
		if ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == ':' || ch == ',' {
			tokens = append(tokens, Token{Text: string(ch), Type: TokenPunctuation, Start: i, End: i + 1})
			i++
			continue
		}

		// Other
		tokens = append(tokens, Token{Text: string(ch), Type: TokenNormal, Start: i, End: i + 1})
		i++
	}

	return tokens
}

func (c *CodeView) tokenizeBash(line string) []Token {
	keywords := map[string]bool{
		"if": true, "then": true, "else": true, "elif": true, "fi": true,
		"for": true, "while": true, "do": true, "done": true, "case": true,
		"esac": true, "in": true, "function": true, "return": true,
		"local": true, "export": true, "readonly": true, "declare": true,
	}

	builtins := map[string]bool{
		"echo": true, "cd": true, "pwd": true, "ls": true, "cat": true,
		"grep": true, "sed": true, "awk": true, "find": true, "xargs": true,
		"read": true, "exit": true, "source": true, "exec": true, "eval": true,
	}

	return c.tokenizeGeneric(line, keywords, builtins, nil, "#", nil, '"', '\'', 0)
}

func (c *CodeView) tokenizeGeneric(line string, keywords, builtins, types map[string]bool, lineComment string, blockComment []string, strQuote1, strQuote2, strQuote3 rune) []Token {
	var tokens []Token
	i := 0
	n := len(line)

	for i < n {
		ch := rune(line[i])

		// Line comment
		if lineComment != "" && i+len(lineComment) <= n && line[i:i+len(lineComment)] == lineComment {
			tokens = append(tokens, Token{Text: line[i:], Type: TokenComment, Start: i, End: n})
			break
		}

		// Whitespace
		if ch == ' ' || ch == '\t' {
			start := i
			for i < n && (line[i] == ' ' || line[i] == '\t') {
				i++
			}
			tokens = append(tokens, Token{Text: line[start:i], Type: TokenNormal, Start: start, End: i})
			continue
		}

		// String
		if ch == strQuote1 || ch == strQuote2 || (strQuote3 != 0 && ch == strQuote3) {
			quote := ch
			start := i
			i++
			for i < n && rune(line[i]) != quote {
				if line[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			if i < n {
				i++
			}
			tokens = append(tokens, Token{Text: line[start:i], Type: TokenString, Start: start, End: i})
			continue
		}

		// Number
		if ch >= '0' && ch <= '9' {
			start := i
			for i < n && ((line[i] >= '0' && line[i] <= '9') || line[i] == '.' || line[i] == 'x' || line[i] == 'X' || (line[i] >= 'a' && line[i] <= 'f') || (line[i] >= 'A' && line[i] <= 'F')) {
				i++
			}
			tokens = append(tokens, Token{Text: line[start:i], Type: TokenNumber, Start: start, End: i})
			continue
		}

		// Identifier
		if unicode.IsLetter(ch) || ch == '_' {
			start := i
			for i < n && (unicode.IsLetter(rune(line[i])) || unicode.IsDigit(rune(line[i])) || line[i] == '_') {
				i++
			}
			word := line[start:i]
			tokenType := TokenNormal
			if keywords[word] {
				tokenType = TokenKeyword
			} else if builtins != nil && builtins[word] {
				tokenType = TokenBuiltin
			} else if types != nil && types[word] {
				tokenType = TokenType_
			}
			tokens = append(tokens, Token{Text: word, Type: tokenType, Start: start, End: i})
			continue
		}

		// Operators and punctuation
		if strings.ContainsRune("+-*/%=<>!&|^~?:;,.(){}[]", ch) {
			tokens = append(tokens, Token{Text: string(ch), Type: TokenOperator, Start: i, End: i + 1})
			i++
			continue
		}

		// Other
		tokens = append(tokens, Token{Text: string(ch), Type: TokenNormal, Start: i, End: i + 1})
		i++
	}

	return tokens
}

func (c *CodeView) getTokenColor(tokenType TokenType) tcell.Color {
	th := c.th()
	switch tokenType {
	case TokenKeyword:
		return th.Accent()
	case TokenString:
		return th.Success()
	case TokenNumber:
		return th.Warning()
	case TokenComment:
		return th.FgDim()
	case TokenFunction:
		return th.Info()
	case TokenType_:
		return th.Accent()
	case TokenBuiltin:
		return th.Warning()
	case TokenOperator, TokenPunctuation:
		return th.FgMuted()
	default:
		return th.Fg()
	}
}

// Draw renders the code view
func (c *CodeView) Draw(screen tcell.Screen) {
	c.Box.DrawForSubclass(screen)
	x, y, width, height := c.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	th := c.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	highlightBg := th.BgLight()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	lineNumStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
	highlightStyle := tcell.StyleDefault.Background(highlightBg)

	// Clear area
	fillRect(screen, x, y, width, height, bgStyle)

	// Calculate line number width
	lineNumWidth := 0
	if c.showLineNumbers {
		lineNumWidth = len(itoa(len(c.lines))) + 2
		c.lineNumberWidth = lineNumWidth
	}

	codeX := x + lineNumWidth
	codeWidth := width - lineNumWidth

	// Draw visible lines
	for i := 0; i < height && c.offsetY+i < len(c.lines); i++ {
		lineIdx := c.offsetY + i
		line := c.lines[lineIdx]
		rowY := y + i

		// Check if line is highlighted
		isHighlighted := lineIdx+1 == c.highlightLine
		rowBg := bgColor
		if isHighlighted {
			rowBg = highlightBg
			// Fill background
			fillLine(screen, x, rowY, width, highlightStyle)
		}

		// Draw line number
		if c.showLineNumbers {
			numStr := itoa(lineIdx + 1)
			numStyle := lineNumStyle
			if isHighlighted {
				numStyle = tcell.StyleDefault.Background(rowBg).Foreground(fgColor)
			}
			for j := 0; j < lineNumWidth-2-len(numStr); j++ {
				screen.SetContent(x+j, rowY, ' ', nil, numStyle)
			}
			for j, r := range numStr {
				screen.SetContent(x+lineNumWidth-2-len(numStr)+j, rowY, r, nil, numStyle)
			}
			screen.SetContent(x+lineNumWidth-1, rowY, ' ', nil, numStyle)
		}

		// Tokenize and draw code
		tokens := c.tokenizeLine(lineIdx, line)
		col := codeX - c.offsetX

		for _, token := range tokens {
			tokenColor := c.getTokenColor(token.Type)
			tokenStyle := tcell.StyleDefault.Background(rowBg).Foreground(tokenColor)

			for _, ch := range token.Text {
				if ch == '\t' {
					// Expand tabs
					spaces := c.tabWidth - ((col - codeX + c.offsetX) % c.tabWidth)
					for s := 0; s < spaces; s++ {
						if col >= codeX && col < codeX+codeWidth {
							screen.SetContent(col, rowY, ' ', nil, tokenStyle)
						}
						col++
					}
				} else {
					if col >= codeX && col < codeX+codeWidth {
						screen.SetContent(col, rowY, ch, nil, tokenStyle)
					}
					col++
				}
			}
		}
	}
}

func (c *CodeView) HandleKey(ev *tcell.EventKey) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, _, _, height := c.GetInnerRect()

	switch ev.Key() {
	case tcell.KeyDown:
		if c.offsetY < len(c.lines)-1 {
			c.offsetY++
		}
		return true
	case tcell.KeyUp:
		if c.offsetY > 0 {
			c.offsetY--
		}
		return true
	case tcell.KeyPgDn:
		c.offsetY += height
		if c.offsetY > len(c.lines)-height {
			c.offsetY = len(c.lines) - height
		}
		if c.offsetY < 0 {
			c.offsetY = 0
		}
		return true
	case tcell.KeyPgUp:
		c.offsetY -= height
		if c.offsetY < 0 {
			c.offsetY = 0
		}
		return true
	case tcell.KeyHome:
		c.offsetY = 0
		return true
	case tcell.KeyEnd:
		c.offsetY = len(c.lines) - height
		if c.offsetY < 0 {
			c.offsetY = 0
		}
		return true
	case tcell.KeyLeft:
		if c.offsetX > 0 {
			c.offsetX--
		}
		return true
	case tcell.KeyRight:
		c.offsetX++
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			if c.offsetY < len(c.lines)-1 {
				c.offsetY++
			}
			return true
		case 'k':
			if c.offsetY > 0 {
				c.offsetY--
			}
			return true
		case 'g':
			c.offsetY = 0
			return true
		case 'G':
			c.offsetY = len(c.lines) - height
			if c.offsetY < 0 {
				c.offsetY = 0
			}
			return true
		case 'h':
			if c.offsetX > 0 {
				c.offsetX--
			}
			return true
		case 'l':
			c.offsetX++
			return true
		}
	}
	return false
}

// GetFieldHeight returns preferred height
func (c *CodeView) GetFieldHeight() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.lines) > 20 {
		return 20
	}
	return len(c.lines) + 2
}

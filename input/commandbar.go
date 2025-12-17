package input

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/theme"
)

// CommandType identifies different types of commands.
// Apps should define their own command types starting from CommandTypeCustom.
type CommandType int

const (
	CommandTypeFilter CommandType = iota
	CommandTypeAction
	CommandTypeSearch
	CommandTypeCustom // Start custom types from here
)

// CommandTypeConfig configures a command type's appearance.
type CommandTypeConfig struct {
	Prompt      string      // e.g., "/", ":", "?"
	Placeholder string      // e.g., "Filter...", "Command..."
	Color       tcell.Color // Prompt color (0 = use accent)
}

// CommandBar is a K9s-style command/filter input bar.
type CommandBar struct {
	*tview.Box
	input       string
	cursorPos   int
	commandType CommandType
	typeConfigs map[CommandType]CommandTypeConfig
	onSubmit    func(cmdType CommandType, input string)
	onCancel    func()
	onChange    func(input string)
	visible     bool
}

// NewCommandBar creates a new command bar.
func NewCommandBar() *CommandBar {
	c := &CommandBar{
		Box:         tview.NewBox(),
		typeConfigs: make(map[CommandType]CommandTypeConfig),
	}

	// Register default configs
	c.Configure(CommandTypeFilter, CommandTypeConfig{
		Prompt:      "/",
		Placeholder: "Filter...",
	})
	c.Configure(CommandTypeAction, CommandTypeConfig{
		Prompt:      ":",
		Placeholder: "Command...",
	})
	c.Configure(CommandTypeSearch, CommandTypeConfig{
		Prompt:      "?",
		Placeholder: "Search...",
	})

	return c
}

// Configure sets the config for a command type.
func (c *CommandBar) Configure(cmdType CommandType, config CommandTypeConfig) *CommandBar {
	c.typeConfigs[cmdType] = config
	return c
}

// Show activates the command bar with the given type.
func (c *CommandBar) Show(cmdType CommandType) *CommandBar {
	c.commandType = cmdType
	c.input = ""
	c.cursorPos = 0
	c.visible = true
	return c
}

// Hide deactivates the command bar.
func (c *CommandBar) Hide() *CommandBar {
	c.visible = false
	c.input = ""
	c.cursorPos = 0
	return c
}

// IsVisible returns whether the bar is visible.
func (c *CommandBar) IsVisible() bool {
	return c.visible
}

// GetInput returns the current input text.
func (c *CommandBar) GetInput() string {
	return c.input
}

// SetInput sets the input text and moves cursor to end.
func (c *CommandBar) SetInput(text string) *CommandBar {
	c.input = text
	c.cursorPos = len(text)
	return c
}

// Clear clears the input.
func (c *CommandBar) Clear() *CommandBar {
	c.input = ""
	c.cursorPos = 0
	return c
}

// GetCommandType returns the current command type.
func (c *CommandBar) GetCommandType() CommandType {
	return c.commandType
}

// SetOnSubmit sets the submit callback (Enter pressed).
func (c *CommandBar) SetOnSubmit(fn func(cmdType CommandType, input string)) *CommandBar {
	c.onSubmit = fn
	return c
}

// SetOnCancel sets the cancel callback (Esc pressed).
func (c *CommandBar) SetOnCancel(fn func()) *CommandBar {
	c.onCancel = fn
	return c
}

// SetOnChange sets the change callback (input changed).
// Useful for filter-as-you-type functionality.
func (c *CommandBar) SetOnChange(fn func(input string)) *CommandBar {
	c.onChange = fn
	return c
}

// Draw renders the command bar.
func (c *CommandBar) Draw(screen tcell.Screen) {
	if !c.visible {
		c.Box.DrawForSubclass(screen, c)
		return
	}

	c.Box.DrawForSubclass(screen, c)

	x, y, width, height := c.GetInnerRect()
	if width < 1 || height < 1 {
		return
	}

	// Get colors
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()

	config := c.typeConfigs[c.commandType]
	promptColor := config.Color
	if promptColor == 0 {
		promptColor = theme.Accent()
	}

	// Draw background
	bgStyle := tcell.StyleDefault.Background(bgColor)
	for col := x; col < x+width; col++ {
		screen.SetContent(col, y, ' ', nil, bgStyle)
	}

	// Draw prompt
	promptStyle := tcell.StyleDefault.Background(bgColor).Foreground(promptColor)
	promptRunes := []rune(config.Prompt)
	for i, r := range promptRunes {
		if x+i < x+width {
			screen.SetContent(x+i, y, r, nil, promptStyle)
		}
	}

	// Draw input or placeholder
	inputX := x + len(promptRunes)
	inputStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	placeholderStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	if c.input == "" && config.Placeholder != "" {
		// Draw placeholder
		for i, r := range config.Placeholder {
			if inputX+i < x+width {
				screen.SetContent(inputX+i, y, r, nil, placeholderStyle)
			}
		}
	} else {
		// Draw input
		inputRunes := []rune(c.input)
		for i, r := range inputRunes {
			if inputX+i < x+width {
				screen.SetContent(inputX+i, y, r, nil, inputStyle)
			}
		}
	}

	// Draw cursor
	cursorX := inputX + c.cursorPos
	if cursorX < x+width {
		cursorStyle := tcell.StyleDefault.Background(fgColor).Foreground(bgColor)
		cursorChar := ' '
		inputRunes := []rune(c.input)
		if c.cursorPos < len(inputRunes) {
			cursorChar = inputRunes[c.cursorPos]
		}
		screen.SetContent(cursorX, y, cursorChar, nil, cursorStyle)
	}
}

// InputHandler handles key events for the command bar.
func (c *CommandBar) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if !c.visible {
			return
		}

		inputRunes := []rune(c.input)

		switch event.Key() {
		case tcell.KeyEnter:
			if c.onSubmit != nil {
				c.onSubmit(c.commandType, c.input)
			}

		case tcell.KeyEscape:
			if c.onCancel != nil {
				c.onCancel()
			}

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if c.cursorPos > 0 {
				inputRunes = append(inputRunes[:c.cursorPos-1], inputRunes[c.cursorPos:]...)
				c.input = string(inputRunes)
				c.cursorPos--
				c.notifyChange()
			}

		case tcell.KeyDelete:
			if c.cursorPos < len(inputRunes) {
				inputRunes = append(inputRunes[:c.cursorPos], inputRunes[c.cursorPos+1:]...)
				c.input = string(inputRunes)
				c.notifyChange()
			}

		case tcell.KeyLeft:
			if c.cursorPos > 0 {
				c.cursorPos--
			}

		case tcell.KeyRight:
			if c.cursorPos < len(inputRunes) {
				c.cursorPos++
			}

		case tcell.KeyHome, tcell.KeyCtrlA:
			c.cursorPos = 0

		case tcell.KeyEnd, tcell.KeyCtrlE:
			c.cursorPos = len(inputRunes)

		case tcell.KeyCtrlU:
			// Clear to start of line
			c.input = string(inputRunes[c.cursorPos:])
			c.cursorPos = 0
			c.notifyChange()

		case tcell.KeyCtrlK:
			// Clear to end of line
			c.input = string(inputRunes[:c.cursorPos])
			c.notifyChange()

		case tcell.KeyCtrlW:
			// Delete word backward
			if c.cursorPos > 0 {
				// Find start of previous word
				pos := c.cursorPos - 1
				for pos > 0 && inputRunes[pos] == ' ' {
					pos--
				}
				for pos > 0 && inputRunes[pos-1] != ' ' {
					pos--
				}
				inputRunes = append(inputRunes[:pos], inputRunes[c.cursorPos:]...)
				c.input = string(inputRunes)
				c.cursorPos = pos
				c.notifyChange()
			}

		case tcell.KeyRune:
			// Insert character at cursor
			r := event.Rune()
			inputRunes = append(inputRunes[:c.cursorPos], append([]rune{r}, inputRunes[c.cursorPos:]...)...)
			c.input = string(inputRunes)
			c.cursorPos++
			c.notifyChange()
		}
	})
}

// notifyChange calls the onChange callback if set.
func (c *CommandBar) notifyChange() {
	if c.onChange != nil {
		c.onChange(c.input)
	}
}

// HasFocus returns true when visible.
func (c *CommandBar) HasFocus() bool {
	return c.visible
}

// Focus makes the command bar focusable when visible.
func (c *CommandBar) Focus(delegate func(tview.Primitive)) {
	if c.visible {
		c.Box.Focus(delegate)
	}
}

// GetPreferredHeight returns the preferred height (1 line).
func (c *CommandBar) GetPreferredHeight() int {
	return 1
}

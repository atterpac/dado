package inline

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
)

var ErrNonInteractive = errors.New("inline: session input is not a terminal")

// SessionOption configures an interactive inline Session.
type SessionOption func(*Session)

// WithSessionInput supplies an input stream. Supplying one explicitly allows
// deterministic non-terminal sessions in tests and embedded applications.
func WithSessionInput(input io.Reader) SessionOption {
	return func(session *Session) {
		if input != nil {
			session.input = input
			session.inputSet = true
		}
	}
}

// WithSessionWidth fixes the render width instead of reading terminal size.
func WithSessionWidth(width int) SessionOption {
	return func(session *Session) { session.width = max(width, 1) }
}

// Session owns raw input mode and drives a Form through an existing Renderer.
// It never enters the alternate screen.
type Session struct {
	renderer *Renderer
	input    io.Reader
	inputSet bool
	width    int
}

func NewSession(renderer *Renderer, options ...SessionOption) *Session {
	session := &Session{renderer: renderer, input: os.Stdin, width: 80}
	for _, option := range options {
		if option != nil {
			option(session)
		}
	}
	return session
}

// Run renders and drives form until submit, cancellation, input failure, or
// context cancellation.
func (s *Session) Run(ctx context.Context, form *Form) (FormResult, error) {
	if s.renderer == nil {
		return nil, errors.New("inline: nil session renderer")
	}
	if form == nil {
		return nil, errors.New("inline: nil form")
	}
	var restore func() error
	if file, ok := s.input.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		state, err := term.MakeRaw(int(file.Fd()))
		if err != nil {
			return nil, err
		}
		restore = func() error { return term.Restore(int(file.Fd()), state) }
		defer restore()
		if width, _, err := term.GetSize(int(file.Fd())); err == nil && width > 0 && s.width == 80 {
			s.width = max(width-1, 1)
		}
	} else if !s.inputSet {
		return nil, ErrNonInteractive
	}

	if err := s.renderer.Render(form.Frame(s.width)); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(s.input)
	type eventResult struct {
		event *tcell.EventKey
		err   error
	}
	for !form.Done() {
		events := make(chan eventResult, 1)
		go func() { event, err := readSessionEvent(reader); events <- eventResult{event: event, err: err} }()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-events:
			if result.err != nil {
				return nil, result.err
			}
			form.HandleKey(result.event)
			if err := s.renderer.Render(form.Frame(s.width)); err != nil {
				return nil, err
			}
		}
	}
	if form.cancelled {
		return nil, ErrFormCancelled
	}
	return form.Result(), nil
}

func readSessionEvent(reader *bufio.Reader) (*tcell.EventKey, error) {
	first, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	switch first {
	case 3:
		return tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl), nil
	case 9:
		return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), nil
	case 13, 10:
		return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil
	case 8, 127:
		return tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone), nil
	case 27:
		return readEscapeEvent(reader)
	}
	if first < utf8.RuneSelf {
		return tcell.NewEventKey(tcell.KeyRune, rune(first), tcell.ModNone), nil
	}
	if err := reader.UnreadByte(); err != nil {
		return nil, err
	}
	value, _, err := reader.ReadRune()
	if err != nil {
		return nil, err
	}
	return tcell.NewEventKey(tcell.KeyRune, value, tcell.ModNone), nil
}

func readEscapeEvent(reader *bufio.Reader) (*tcell.EventKey, error) {
	if reader.Buffered() == 0 {
		return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil
	}
	next, err := reader.ReadByte()
	if err != nil {
		return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil
	}
	if next != '[' {
		return tcell.NewEventKey(tcell.KeyRune, rune(next), tcell.ModAlt), nil
	}
	code, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	switch code {
	case 'A':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), nil
	case 'B':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil
	case 'C':
		return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), nil
	case 'D':
		return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone), nil
	case 'H':
		return tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone), nil
	case 'F':
		return tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone), nil
	case 'Z':
		return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift), nil
	case '3':
		if tail, _ := reader.ReadByte(); tail == '~' {
			return tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone), nil
		}
	}
	return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil
}

package inline

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
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
	return func(session *Session) {
		session.width = max(width, 1)
		session.widthSet = true
	}
}

// Session owns raw input mode and drives a Form through an existing Renderer.
// It never enters the alternate screen.
type Session struct {
	renderer *Renderer
	input    io.Reader
	inputSet bool
	width    int
	widthSet bool
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
func (s *Session) Run(ctx context.Context, form *Form) (result FormResult, runErr error) {
	if s.renderer == nil {
		return nil, errors.New("inline: nil session renderer")
	}
	if form == nil {
		return nil, errors.New("inline: nil form")
	}
	if err := form.validateStructure(); err != nil {
		return nil, err
	}

	if file, ok := s.input.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		state, err := term.MakeRaw(int(file.Fd()))
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := term.Restore(int(file.Fd()), state); err != nil && runErr == nil {
				result = nil
				runErr = fmt.Errorf("inline: restore terminal: %w", err)
			}
		}()
		if width, _, err := term.GetSize(int(file.Fd())); err == nil && width > 0 && !s.widthSet {
			s.width = max(width-1, 1)
		}
	} else if !s.inputSet {
		return nil, ErrNonInteractive
	}

	if err := s.renderer.Render(form.Frame(s.width)); err != nil {
		return nil, err
	}
	reader := newSessionEventReader(s.input)
	for !form.Done() {
		event, err := reader.readEvent(ctx)
		if err != nil {
			return nil, err
		}
		form.HandleKey(event)
		if err := s.renderer.Render(form.Frame(s.width)); err != nil {
			return nil, err
		}
	}
	if form.cancelled {
		if form.interrupted {
			return nil, ErrFormInterrupted
		}
		return nil, ErrFormCancelled
	}
	return form.Result(), nil
}

const escapeSequenceTimeout = 35 * time.Millisecond

type sessionEventReader struct {
	reader *bufio.Reader
	file   *os.File
}

func newSessionEventReader(input io.Reader) *sessionEventReader {
	reader := &sessionEventReader{reader: bufio.NewReader(input)}
	reader.file, _ = input.(*os.File)
	return reader
}

func (r *sessionEventReader) readEvent(ctx context.Context) (*tcell.EventKey, error) {
	first, err := r.readByte(ctx, 0)
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
		return r.readEscapeEvent(ctx)
	}
	if first < utf8.RuneSelf {
		return tcell.NewEventKey(tcell.KeyRune, rune(first), tcell.ModNone), nil
	}
	value, err := r.readRune(ctx, first)
	if err != nil {
		return nil, err
	}
	return tcell.NewEventKey(tcell.KeyRune, value, tcell.ModNone), nil
}

func (r *sessionEventReader) readEscapeEvent(ctx context.Context) (*tcell.EventKey, error) {
	next, err := r.readByte(ctx, escapeSequenceTimeout)
	if errors.Is(err, errSessionInputTimeout) || errors.Is(err, io.EOF) {
		return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil
	}
	if err != nil {
		return nil, err
	}
	if next != '[' {
		value := rune(next)
		if next >= utf8.RuneSelf {
			value, err = r.readRune(ctx, next)
			if err != nil {
				return nil, err
			}
		}
		return tcell.NewEventKey(tcell.KeyRune, value, tcell.ModAlt), nil
	}
	code, err := r.readByte(ctx, escapeSequenceTimeout)
	if err != nil {
		if errors.Is(err, errSessionInputTimeout) || errors.Is(err, io.EOF) {
			return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil
		}
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
		if tail, _ := r.readByte(ctx, escapeSequenceTimeout); tail == '~' {
			return tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone), nil
		}
	}
	return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil
}

func (r *sessionEventReader) readByte(ctx context.Context, timeout time.Duration) (byte, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	if r.reader.Buffered() > 0 || r.file == nil {
		return r.reader.ReadByte()
	}
	ready, err := waitForSessionInput(ctx, r.file, timeout)
	if err != nil {
		return 0, err
	}
	if !ready {
		return 0, errSessionInputTimeout
	}
	return r.reader.ReadByte()
}

func (r *sessionEventReader) readRune(ctx context.Context, first byte) (rune, error) {
	encoded := []byte{first}
	for !utf8.FullRune(encoded) && len(encoded) < utf8.UTFMax {
		next, err := r.readByte(ctx, 0)
		if err != nil {
			return 0, err
		}
		encoded = append(encoded, next)
	}
	value, _ := utf8.DecodeRune(encoded)
	return value, nil
}

//go:build windows

package inline

import (
	"context"
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

var errSessionInputTimeout = errors.New("inline: session input timeout")

func waitForSessionInput(ctx context.Context, file *os.File, timeout time.Duration) (bool, error) {
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		wait := 50 * time.Millisecond
		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return false, nil
			}
			wait = min(wait, remaining)
		}
		event, err := windows.WaitForSingleObject(windows.Handle(file.Fd()), uint32(max(wait.Milliseconds(), 1)))
		if err != nil {
			return false, err
		}
		switch event {
		case windows.WAIT_OBJECT_0:
			return true, nil
		case uint32(windows.WAIT_TIMEOUT):
			continue
		default:
			return false, errors.New("inline: unexpected terminal wait result")
		}
	}
}

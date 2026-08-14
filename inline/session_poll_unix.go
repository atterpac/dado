//go:build unix

package inline

import (
	"context"
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

var errSessionInputTimeout = errors.New("inline: session input timeout")

func waitForSessionInput(ctx context.Context, file *os.File, timeout time.Duration) (bool, error) {
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	pollFD := []unix.PollFd{{Fd: int32(file.Fd()), Events: unix.POLLIN | unix.POLLHUP}}
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
		n, err := unix.Poll(pollFD, max(int(wait.Milliseconds()), 1))
		if err != nil && err != unix.EINTR {
			return false, err
		}
		if n > 0 {
			return true, nil
		}
	}
}

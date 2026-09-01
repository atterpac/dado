//go:build !unix && !windows

package inline

import (
	"context"
	"errors"
	"os"
	"time"
)

var errSessionInputTimeout = errors.New("inline: session input timeout")

func waitForSessionInput(ctx context.Context, _ *os.File, timeout time.Duration) (bool, error) {
	if timeout <= 0 {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			return true, nil
		}
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-timer.C:
		return false, nil
	}
}

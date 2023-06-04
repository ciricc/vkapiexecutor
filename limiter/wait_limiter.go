package limiter

import "context"

type WaitLimiter interface {
	// Wait executes each request
	Wait(ctx context.Context) error
}

var DefaultLimiter WaitLimiter = &emptyLimiter{}

type emptyLimiter struct{}

func (e *emptyLimiter) Wait(ctx context.Context) error {
	return nil
}

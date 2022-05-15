package limiter_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ciricc/vkapiexecutor/limiter"
	"github.com/ciricc/vkapiexecutor/request"
)

func TestLimiter(t *testing.T) {
	t.Run("rps test", func(t *testing.T) {
		// 1 requests per second on 10 seconds duration calls
		rps := 1
		runDurationInSeconds := 10
		expectedCallsCount := rps * runDurationInSeconds

		limiter := limiter.New(rps, 10*time.Minute, time.Hour)
		handleFunc := limiter.Handle()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runDurationInSeconds)*time.Second)
		defer cancel()

		var completedRequests int32 = 0
		var ctxDone = false

		for i := 0; i < 10000; i++ {
			reqParams := request.NewParams()
			reqParams.AccessToken("abc")

			req := request.New()
			req.Params(reqParams)

			handleFunc(func(ctx context.Context, _ *request.Request) error {
				if ctxDone {
					atomic.AddInt32(&completedRequests, 1)
				}
				return nil
			}, ctx, req)
		}

		<-ctx.Done()
		ctxDone = true

		if completedRequests > int32(expectedCallsCount) {
			t.Errorf("Requests made in one minute: %d, but expected: %d", completedRequests, expectedCallsCount)
		}
	})

	t.Run("no token rps", func(t *testing.T) {
		rps := 1
		runDurationInSeconds := 5
		expectedCallsCount := 10000

		limiter := limiter.New(rps, 10*time.Minute, time.Hour)
		handleFunc := limiter.Handle()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runDurationInSeconds)*time.Second)
		defer cancel()

		var completedRequests int32 = 0
		var ctxDone = false

		for i := 0; i < 10000; i++ {
			req := request.New()
			handleFunc(func(ctx context.Context, _ *request.Request) error {
				if ctxDone {
					atomic.AddInt32(&completedRequests, 1)
				}
				return nil
			}, ctx, req)
		}

		<-ctx.Done()
		ctxDone = true

		if completedRequests > int32(expectedCallsCount) {
			t.Errorf("Requests made in one minute: %d, but expected: %d", completedRequests, expectedCallsCount)
		}
	})
}

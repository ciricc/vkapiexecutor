package limiter_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runDurationInSeconds)*time.Second)
		defer cancel()

		req := request.New()
		req.Method("users.get")

		params := request.NewParams()
		params.AccessToken("abc")

		req.Params(params)

		exec := executor.New()
		exec.HttpClient = &http.Client{
			Transport: limiter,
		}

		var completedRequests int32 = 0

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if completedRequests > int32(expectedCallsCount) {
					t.Errorf("Requests made: %d, but expected: %d", completedRequests, expectedCallsCount)
					return
				}

				exec.DoRequest(req)
				completedRequests += 1
			}
		}
	})

	t.Run("no token rps", func(t *testing.T) {
		rps := 1
		runDurationInSeconds := 5
		expectedCallsCount := 20

		limiter := limiter.New(rps, 10*time.Minute, time.Hour)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runDurationInSeconds)*time.Second)
		defer cancel()

		var completedRequests int32 = 0

		req := request.New()
		req.Method("users.get")

		exec := executor.New()
		exec.HttpClient = &http.Client{
			Transport: limiter,
		}

		for {
			select {
			case <-ctx.Done():
				if completedRequests < int32(expectedCallsCount) {
					t.Errorf("Requests made: %d, but expected minimum: %d", completedRequests, expectedCallsCount)
					return
				}
				return
			default:
				exec.DoRequest(req)
				completedRequests += 1
			}
		}
	})
}

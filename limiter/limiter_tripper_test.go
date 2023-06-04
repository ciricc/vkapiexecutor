package limiter_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/limiter"
	"github.com/ciricc/vkapiexecutor/request"
)

func TestLimiter(t *testing.T) {
	t.Run("rps test", func(t *testing.T) {
		// 1 requests per second on 10 seconds duration calls
		testCases := []struct {
			rps                 int
			runDurationInSecods int
			params              *request.Params
		}{
			{
				rps:                 1,
				runDurationInSecods: 10,
				params:              request.NewParams(),
			},
			{
				rps:                 1,
				runDurationInSecods: 10,
				params:              request.NewParams(),
			},
		}

		// Set variant of tokens
		testCases[0].params.AccessToken("use_access_token")
		testCases[1].params.AnonymousToken("use_anonymous_token")

		wg := sync.WaitGroup{}
		for _, test := range testCases {
			test := test
			wg.Add(1)
			go func() {
				defer wg.Done()
				expectedCallsCount := test.rps * test.runDurationInSecods
				limiter := limiter.New(test.rps, 10*time.Minute, time.Hour)

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(test.runDurationInSecods)*time.Second)
				defer cancel()

				req := request.New()
				req.Method("users.get")
				req.Params(test.params)

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
							t.Errorf("Requests made: %d, but expected: %d, test_case=%v params=%v", completedRequests, expectedCallsCount, test, *test.params)
							return
						}

						exec.DoRequest(req)
						completedRequests += 1
					}
				}
			}()
		}
		wg.Wait()
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

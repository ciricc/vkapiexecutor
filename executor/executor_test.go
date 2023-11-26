package executor_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/request"
	"github.com/ciricc/vkapiexecutor/response"
)

type MessagepackParser struct {
	Parsed bool
}

func (v *MessagepackParser) Parse(httpResponse *http.Response) (response.Response, error) {
	v.Parsed = true
	return response.NewUnknown(httpResponse), nil
}

type CustomHttpRoundTripper struct {
	rt http.RoundTripper
	t  *testing.T
}

func (v *CustomHttpRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return v.rt.RoundTrip(req)
}

func solveCaptchaExample(captchaImg string) (string, error) {
	return "captcha_key_code", nil
}

func TestExecutor(t *testing.T) {
	apiExecutor := executor.New()

	t.Run("api endpoint", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		res, err := apiExecutor.DoRequestCtx(context.Background(), req)
		if res == nil && err != nil {
			t.Errorf("http error: %s", err)
		}

		expectedUrl := "https://api.vk.com/method/users.get"
		if res.HttpResponse().Request.URL.String() != expectedUrl {
			t.Errorf("expected api request url: %q but got %q", expectedUrl, res.HttpResponse().Request.URL.String())
		}
	})

	t.Run("close request", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		res, err := apiExecutor.DoRequestCtx(context.Background(), req)
		if err != nil && res == nil {
			t.Error(err)
		}
		if _, err := res.HttpResponse().Body.Read(nil); err == nil {
			t.Errorf("response body not closed")
		}
	})

	t.Run("response is not nil when http error is nil", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		res, err := apiExecutor.DoRequestCtx(context.Background(), req)
		if err != nil {
			switch errBody := err.(type) {
			case *response.Error:
				if res == nil {
					t.Errorf("response nil, but http response is got %v", errBody)
				}
			}
		}
	})

	t.Run("do not execute nil request", func(t *testing.T) {
		_, err := apiExecutor.DoRequestCtx(context.Background(), nil)
		if err == nil {
			t.Errorf("request executed, but should not be")
		}
	})

	t.Run("execute timeout", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		res, _ := apiExecutor.DoRequestCtx(ctx, req)
		if res != nil {
			t.Errorf("expected timeout error")
		}
	})

	t.Run("get request from context", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		res, err := apiExecutor.DoRequestCtx(context.Background(), req)
		if err != nil {
			if _, ok := err.(*response.Error); !ok {
				t.Error(err)
			}
		}

		ctxReq := executor.GetRequest(res.Context())
		if ctxReq != req {
			t.Errorf("requests different\norigin: %q\n\ncontextValue: %q", req, ctxReq)
		}
	})

	t.Run("custom response parser", func(t *testing.T) {

		req := request.New()
		req.Method("users.get.msgpack")

		apiExecutorCustomParser := executor.New()
		msgPackParser := MessagepackParser{}
		apiExecutorCustomParser.ResponseParser = &msgPackParser

		res, err := apiExecutorCustomParser.DoRequest(req)

		if err != nil && res == nil {
			t.Error("http response error: %w", err)
		}

		if !msgPackParser.Parsed {
			t.Error("not parsed by custom parser")
		}
	})

	t.Run("concurrently use one request in one executor", func(t *testing.T) {
		req := request.New()
		exec := executor.New()

		req.Method("users.get") // it's ok, because nothing changes in concurrency

		wg := sync.WaitGroup{}
		for i := 0; i < 25; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// WARNING
				req.Method("users.get") // DO NOT do this in concurrently system calls with different methods without mutexes

				exec.DoRequest(req) // it's ok, will return response only for this request execution

			}()
		}

		wg.Wait()
	})

	t.Run("response middleware returning error", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")
		exec := executor.New()
		returnErr := fmt.Errorf("test")
		exec.ApiResponseHook(func(next executor.ApiResponseNextHook, res response.Response) error {
			return returnErr
		})
		_, err := exec.DoRequest(req)

		if err != returnErr {
			t.Errorf("got error: %s", returnErr)
		}
	})

	t.Run("http rewrite response body", func(t *testing.T) {
		req := request.New()
		req.Method("some.not.found.method")

		exec := executor.New()

		exec.HttpResponseHook(func(next executor.HttpResponseNextHook, res *http.Response) error {
			req := executor.GetRequest(res.Request.Context())
			if req != nil {
				if req.GetMethod() == "some.not.found.method" {
					res.Body.Close() // important to close body !!
					res.Body = io.NopCloser(strings.NewReader(`{"response": true}`))
				}
			}

			return next(res)
		})

		_, err := exec.DoRequest(req)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("use default http client instead of creating new", func(t *testing.T) {
		exec := executor.New()
		if exec.HttpClient != http.DefaultClient {
			t.Errorf("not using default http client")
		}
	})

	t.Run("change default response parser", func(t *testing.T) {
		rp := executor.DefaultResponseParser
		defer func() {
			executor.DefaultResponseParser = rp
		}()
		executor.DefaultResponseParser = &MessagepackParser{}
		exec := executor.New()
		if exec.ResponseParser != executor.DefaultResponseParser {
			t.Errorf("not using default response parser")
		}
	})
}

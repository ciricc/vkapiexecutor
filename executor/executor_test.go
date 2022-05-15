package executor_test

import (
	"context"
	"fmt"
	"net/http"
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
		ctxReq, err := request.FromContext(res.Context())
		if err != nil {
			t.Errorf("get request from context: %s", err)
		}
		if ctxReq.String() != req.String() {
			t.Errorf("requests different\norigin: %q\n\ncontextValue: %q", req, ctxReq)
		}
	})

	t.Run("test api middlewares", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		madeTest2 := false
		madeTest1 := false

		apiExecutor.HandleApiRequest(func(next executor.ApiRequestHandlerNext, ctx context.Context, req *request.Request) error {
			return next(ctx, req)
		})

		apiExecutor.HandleApiRequest(func(next executor.ApiRequestHandlerNext, ctx context.Context, req *request.Request) error {
			madeTest1 = true
			return next(ctx, req)
		})

		apiExecutor.HandleApiRequest(func(next executor.ApiRequestHandlerNext, ctx context.Context, req *request.Request) error {
			if madeTest1 {
				t.Errorf("api middleware_1 earlier than first")
			}

			madeTest2 = true
			req.Method("middleware.test")
			return next(ctx, req)
		})

		res, err := apiExecutor.DoRequest(req)
		if err != nil && res == nil {
			t.Error(err)
		}

		if !madeTest1 {
			t.Errorf("middleware_1 not called")
		}

		if !madeTest2 {
			t.Errorf("middleware_2 not called")
		}

		if req.GetMethod() != "middleware.test" {
			t.Errorf("middleware_2 nothing changed")
		}

		apiExecutor.ResetApiRequestHandlers()
	})

	t.Run("clear api middlwares", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		callDeletedMiddleware := false
		apiExecutor.HandleApiRequest(func(next executor.ApiRequestHandlerNext, ctx context.Context, req *request.Request) error {
			callDeletedMiddleware = true
			return next(ctx, req)
		})

		apiExecutor.ResetApiRequestHandlers()
		res, err := apiExecutor.DoRequest(req)
		if res == nil && err != nil {
			t.Error(err)
		}

		if callDeletedMiddleware {
			t.Errorf("Call cleaned middleware")
		}

	})

	t.Run("middleware returns error", func(t *testing.T) {
		middlewareError := fmt.Errorf("middleware error")

		req := request.New()
		req.Method("users.get")

		apiExecutorNew := executor.New()

		apiExecutorNew.HandleApiRequest(func(next executor.ApiRequestHandlerNext, ctx context.Context, req *request.Request) error {
			return middlewareError
		})

		_, err := apiExecutorNew.DoRequest(req)
		if err != middlewareError {
			t.Errorf("different errors: \n\norigin: %v, \n\nresult: %v", err, middlewareError)
		}
	})

	t.Run("middleware for http request", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		madeTest2 := false
		madeTest1 := false

		apiExecutor.HandleHttpRequest(func(next executor.HttpRequestHandlerNext, req *http.Request) error {
			return next(req)
		})

		apiExecutor.HandleHttpRequest(func(next executor.HttpRequestHandlerNext, req *http.Request) error {
			madeTest1 = true
			return next(req)
		})

		apiExecutor.HandleHttpRequest(func(next executor.HttpRequestHandlerNext, req *http.Request) error {
			if madeTest1 {
				t.Errorf("http middleware_1 earlier than first")
			}

			madeTest2 = true
			return next(req)
		})

		res, err := apiExecutor.DoRequest(req)
		if err != nil && res == nil {
			t.Error(err)
		}

		if !madeTest1 {
			t.Errorf("first added http middleware not called")
		}

		if !madeTest2 {
			t.Errorf("second added http middleware not called")
		}

		apiExecutor.ResetHttpRequestHandlers()
	})

	t.Run("clear http request middleware handlers", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")
		callHttpMiddleware := false

		apiExecutor.HandleHttpRequest(func(next executor.HttpRequestHandlerNext, req *http.Request) error {
			callHttpMiddleware = true
			return next(req)
		})

		apiExecutor.ResetHttpRequestHandlers()

		res, err := apiExecutor.DoRequest(req)
		if err != nil && res == nil {
			t.Error(err)
		}

		if callHttpMiddleware {
			t.Errorf("deleted http middleware called")
		}
	})

	t.Run("middleware changes http request credentials", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		defer apiExecutor.ResetHttpRequestHandlers()

		apiExecutor.HandleHttpRequest(func(next executor.HttpRequestHandlerNext, req *http.Request) error {
			req.URL.Host = "api.vkontakte.ru"
			return next(req)
		})

		res, err := apiExecutor.DoRequest(req)
		if err != nil && res == nil {
			t.Error(err)
		}

		responseLocation := res.HttpResponse().Request.URL
		if responseLocation.Host != "api.vkontakte.ru" {
			t.Errorf("http middleware nothing changed")
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

	t.Run("middleware which stops request", func(t *testing.T) {
		params := request.NewParams()

		req := request.New()
		req.Params(params)

		exec := executor.New()

		exec.HandleApiRequest(func(next executor.ApiRequestHandlerNext, ctx context.Context, req *request.Request) error {
			if req.GetMethod() == "" || req.GetParams().GetAccessToken() == "" {
				req.Block(true)
			}
			return next(ctx, req)
		})

		res, err := exec.DoRequest(req)

		if err != nil && res != nil {
			t.Errorf("not blocked request")
		}
	})

	t.Run("concurrently use one request in one executor", func(t *testing.T) {
		req := request.New()
		wg := sync.WaitGroup{}
		exec := executor.New()
		for i := 0; i < 25; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req.Method("users.get")
				exec.DoRequest(req) // yes
			}()
		}
		wg.Wait()
	})
}

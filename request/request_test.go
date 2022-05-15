package request_test

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/ciricc/vkapiexecutor/request"
)

func TestRequest(t *testing.T) {

	t.Run("test request method changes", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")
		if req.GetMethod() != "users.get" {
			t.Errorf("not changed method")
		}
	})

	t.Run("test request params changing", func(t *testing.T) {
		req := request.New()

		params := request.NewParams()
		req.Params(params)

		if req.GetParams() != params {
			t.Errorf("not changed params")
		}
	})

	t.Run("serializing request in string", func(t *testing.T) {
		req := request.New()
		log.Println(req.String())
		if len(req.String()) == 0 {
			t.Errorf("serialized request empty")
		}
	})

	t.Run("request setting context value correctly", func(t *testing.T) {
		req := request.New()
		ctx := context.Background()
		ctx = req.SetContextValue(ctx)

		ctxReq, err := request.FromContext(ctx)
		if err != nil {
			t.Error(err)
		}

		if req != ctxReq {
			t.Error("requests are not the same")
		}
	})

	t.Run("request url correctly built", func(t *testing.T) {
		req := request.New()
		req.Method("users.get")

		url, err := req.GetRequestUrl()
		if err != nil {
			t.Error(err)
		}

		expectedUrl := "https://api.vk.com/method/users.get"
		if url.String() != expectedUrl {
			t.Errorf("expected url: %q\ngot: %q", expectedUrl, url)
		}

		req.Method("/users.get")

		url, err = req.GetRequestUrl()
		if err != nil {
			t.Error(err)
		}

		if url.String() != expectedUrl {
			t.Errorf("expected url: %q\ngot: %q", expectedUrl, url)
		}

		req.Method("/users.get/test")
		expectedUrl = "https://api.vk.com/method/users.get/test"

		url, err = req.GetRequestUrl()
		if err != nil {
			t.Error(err)
		}

		if url.String() != expectedUrl {
			t.Errorf("expected url: %q\ngot: %q", expectedUrl, url)
		}
	})

	t.Run("set custom request headers", func(t *testing.T) {
		req := request.New()
		expectedUserAgent := "VKAndroidApp/5.17-2625 (Android 7.1.1; SDK 25; armeabi-v7a; samsung SM-G977N; ru; 1600x900)"

		req.Headers(http.Header{
			"User-Agent": {expectedUserAgent},
		})

		if req.GetHeaders().Get("user-agent") != expectedUserAgent {
			t.Errorf("header value expected: %q\ngot: %q", expectedUserAgent, req.GetHeaders().Get("user-agent"))
		}
	})

	t.Run("built http request same with built-in methods", func(t *testing.T) {
		req := request.New()

		httpReq, err := req.HttpRequest("GET")
		if err != nil {
			t.Error(err)
		}

		httpReqBuiltIn, err := req.HttpRequestGet()
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(httpReqBuiltIn, httpReq) {
			t.Errorf("http get requests different: custom: %v\nbuilt-in: %v", httpReq, httpReqBuiltIn)
		}

		httpReq, err = req.HttpRequest("POST")
		if err != nil {
			t.Error(err)
		}

		httpReqBuiltIn, err = req.HttpRequestPost()
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(httpReqBuiltIn, httpReq) {
			t.Errorf("http post requests different: custom: %v\nbuilt-in: %v", httpReq, httpReqBuiltIn)
		}

	})

	t.Run("correct built get http request", func(t *testing.T) {

		params := request.NewParams()
		params.AccessToken("abc")

		req := request.New()

		req.Method("users.get")
		req.Headers(http.Header{
			"Authorization": {"1"},
		})
		req.Params(params)

		httpReq, err := req.HttpRequestGet()
		if err != nil {
			t.Error(err)
		}

		expectedReqUrl, err := url.Parse("https://api.vk.com/method/users.get?access_token=abc&device_id=&lang=en&v=5.131")
		if err != nil {
			t.Error(err)
		}

		expectedHttpRequest := &http.Request{
			Method: "GET",
			URL:    expectedReqUrl,
			Header: http.Header{
				"Content-Type":  {"application/x-www-form-urlencoded"},
				"Authorization": {"1"},
			},
		}

		if !reflect.DeepEqual(httpReq, expectedHttpRequest) {
			t.Errorf("built http request different with expected\nexpected: %v\nreal: %v", expectedHttpRequest, httpReq)
		}
	})

	t.Run("correct built post http request", func(t *testing.T) {

		params := request.NewParams()
		params.AccessToken("abc")

		req := request.New()

		req.Method("users.get")
		req.Headers(http.Header{
			"Authorization": {"1"},
		})

		req.Params(params)

		httpReq, err := req.HttpRequestPost()
		if err != nil {
			t.Error(err)
		}

		expectedReqUrl, err := url.Parse("https://api.vk.com/method/users.get")
		if err != nil {
			t.Error(err)
		}

		expectedHttpRequest := &http.Request{
			Method: "POST",
			URL:    expectedReqUrl,
			Header: http.Header{
				"Content-Type":  {"application/x-www-form-urlencoded"},
				"Authorization": {"1"},
			},
			Body: io.NopCloser(bytes.NewBuffer([]byte(
				"access_token=abc&device_id=&lang=en&v=5.131",
			))),
		}

		if !reflect.DeepEqual(httpReq, expectedHttpRequest) {
			t.Errorf("built http request different with expected\nexpected: %v\nreal: %v", expectedHttpRequest, httpReq)
		}
	})

	t.Run("append headers", func(t *testing.T) {
		req := request.New()

		req.AppendHeaders(http.Header{
			"Authorization": {"1"},
		})

		req.AppendHeaders(http.Header{
			"User-Agent": {"2"},
		})

		if req.GetHeaders().Get("authorization") != "1" || req.GetHeaders().Get("user-agent") != "2" {
			t.Errorf("append headers not all: %v", req.GetHeaders())
		}
	})

	t.Run("append content-type will not work", func(t *testing.T) {
		req := request.New()
		req.AppendHeaders(http.Header{
			"Content-Type": {"application/json"},
		})

		if req.GetHeaders().Get("content-type") == "application/json" {
			t.Errorf("appended content-type header: %q", req.GetHeaders().Get("content-type"))
		}
	})

	t.Run("rewrite headers will not remove content-type", func(t *testing.T) {
		req := request.New()

		req.Headers(http.Header{
			"Content-Type": {"application/json"},
			"User-Agent":   {"123"},
		})

		if req.GetHeaders().Get("content-type") == "application/json" {
			t.Errorf("removed content-type: %q", req.GetHeaders().Get("content-type"))
		}

		if req.GetHeaders().Get("User-Agent") != "123" {
			t.Errorf("not added authorization header")
		}
	})

	t.Run("nil values set", func(t *testing.T) {
		req := request.New()
		req.Params(nil)

		if req.GetParams() != nil {
			t.Errorf("params is not nil")
		}

		req.AppendHeaders(nil) // will be ok, cause in range cycle
		req.Headers(nil)
		req.HttpRequest("GET") // check no panic

		if req.GetHeaders() == nil {
			t.Errorf("headers is nil")
		}
	})
}

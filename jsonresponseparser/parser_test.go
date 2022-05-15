package jsonresponseparser_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/ciricc/vkapiexecutor/jsonresponseparser"
	"github.com/ciricc/vkapiexecutor/response"
	"github.com/ciricc/vkapiexecutor/responseparser"
)

type JsonResponseTestCase struct {
	ResultBody    string
	HasError      bool
	HasParseError bool
	Tag           string
}

func TestParser(t *testing.T) {
	httpRes, err := http.Get("https://api.vk.com/method/users.get")
	if err != nil {
		t.Error(err)
	}

	var responseParser responseparser.Parser = &jsonresponseparser.JsonResponseParser{}
	res, err := responseParser.Parse(httpRes)

	if err != nil {
		t.Errorf("parse error: %s", err)
	}

	t.Run("response has error", func(t *testing.T) {
		if res.Error() == nil {
			t.Errorf("response has no errors")
		}
	})

	vkApiTestCases := []JsonResponseTestCase{
		{
			ResultBody:    `{"error":{"error_msg":"Captcha needed", "error_code":14, "captcha_sid":"1", "captcha_img":"https://vk.com/captcha.php?sid=1"}}`,
			HasError:      true,
			HasParseError: false,
			Tag:           "captcha error",
		},
		{
			ResultBody:    `{"response":1}`,
			HasError:      false,
			Tag:           "response ok",
			HasParseError: false,
		},
		{
			ResultBody:    `{"error":{"error_msg":"Validation required", "error_code": 17, "redirect_uri":"redirect_uri"}}`,
			HasError:      true,
			HasParseError: false,
			Tag:           "validation error",
		},
		{
			ResultBody:    `EOF`,
			HasError:      false,
			HasParseError: true,
			Tag:           "something wrong with vk api",
		},
	}
	for _, testCase := range vkApiTestCases {
		customResponse := http.Response{
			Body: io.NopCloser(bytes.NewBuffer([]byte(testCase.ResultBody))),
		}

		res, err = responseParser.(*jsonresponseparser.JsonResponseParser).Parse(&customResponse)
		if err != nil {
			if !testCase.HasParseError {
				t.Error(err)
			}
		} else {
			if testCase.HasParseError {
				t.Errorf("expected parse error but error is nil.\ntest case: %q", testCase.Tag)
			}
		}

		if res.Error() == nil && testCase.HasError {
			t.Errorf("expected error on test case, but got nil. Test case: %s", testCase.Tag)
		} else if res.Error() != nil && !testCase.HasError {
			t.Errorf("expected error nil on test case, but got error: %q. Test case: %s", res.Error(), testCase.Tag)
		}
	}

	t.Run("validation error fields test", func(t *testing.T) {
		res, err := responseParser.Parse(&http.Response{
			Body: io.NopCloser(bytes.NewBuffer([]byte(`{"error":{"error_msg":"Validation required", "error_code": 17, "redirect_uri":"redirect_uri"}}`))),
		})
		if err != nil {
			t.Error(err)
		}
		apiError := res.Error()
		if apiError == nil {
			t.Errorf("no error found")
		}
		switch errorObject := apiError.(type) {
		case *response.Error:
			if errorObject.RedirectUri != "redirect_uri" {
				t.Errorf("not found redirect uri string in error\ngot: %q, expected: %q", errorObject.RedirectUri, "redirect_uri")
			}
		default:
			t.Errorf("unknown error type")
		}
	})

	t.Run("captcha error fields test", func(t *testing.T) {
		res, err := responseParser.(*jsonresponseparser.JsonResponseParser).Parse(&http.Response{
			Body: io.NopCloser(bytes.NewBuffer([]byte(`{"error":{"error_msg":"Captcha needed", "error_code":14, "captcha_sid":"1", "captcha_img":"https://vk.com/captcha.php?sid=1"}}`))),
		})
		if err != nil {
			t.Error(err)
		}
		apiError := res.Error()
		if apiError == nil {
			t.Errorf("no error found")
		}
		switch errorObject := apiError.(type) {
		case *response.Error:
			expectedCaptchaImg := "https://vk.com/captcha.php?sid=1"
			if errorObject.CaptchaImg != expectedCaptchaImg {
				t.Errorf("captcha img different \ngot: %q, expected: %q", errorObject.CaptchaImg, expectedCaptchaImg)
			}
			expectedCaptchaSid := "1"
			if errorObject.CaptchaSid != expectedCaptchaSid {
				t.Errorf("captcha sid different \ngot: %q, expected: %q", errorObject.CaptchaSid, expectedCaptchaSid)
			}
		default:
			t.Errorf("unknown error type")
		}
	})

	t.Run("error global fields test", func(t *testing.T) {
		res, err := responseParser.(*jsonresponseparser.JsonResponseParser).Parse(&http.Response{
			Body: io.NopCloser(bytes.NewBuffer([]byte(`{"error":{"error_msg":"Captcha needed", "error_code":14, "captcha_sid":"1", "captcha_img":"https://vk.com/captcha.php?sid=1"}}`))),
		})
		if err != nil {
			t.Error(err)
		}
		apiError := res.Error()
		if apiError == nil {
			t.Errorf("no error found")
		}
		switch errorObject := apiError.(type) {
		case *response.Error:
			expectedErrorCode := 14
			if errorObject.IntCode() != expectedErrorCode {
				t.Errorf("error code different \ngot: %d, expected: %d", errorObject.IntCode(), expectedErrorCode)
			}
			expectedErrorMessage := "Captcha needed"
			if errorObject.Error() != expectedErrorMessage {
				t.Errorf("error message different \ngot: %q, expected: %q", errorObject.Error(), expectedErrorMessage)
			}
		default:
			t.Errorf("unknown error type")
		}
	})
}

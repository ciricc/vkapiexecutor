package jsonresponseparser_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ciricc/vkapiexecutor/jsonresponseparser"
	"github.com/ciricc/vkapiexecutor/response"
	"github.com/ciricc/vkapiexecutor/responseparser"
	"github.com/stretchr/testify/require"
)

type JsonResponseTestCase struct {
	ResultBody    string
	HasError      bool
	HasParseError bool
	Tag           string
}

func TestParser(t *testing.T) {
	t.Parallel()

	var responseParser responseparser.Parser = &jsonresponseparser.JsonResponseParser{}

	t.Run("simple parse test cases", func(t *testing.T) {
		t.Parallel()
		vkApiTestCases := []JsonResponseTestCase{
			{
				ResultBody: `{"error":{"error_msg":"Captcha needed", "error_code":14, "captcha_sid":"1", "captcha_img":` +
					`"https://vk.com/captcha.php?sid=1"}}`,
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
			{
				ResultBody: `{"response":[false],"execute_errors":[{"method":"messages.send","error_code":900,"error_msg":` +
					`"Can't send messages for users from blacklist"}]}`,
				HasError:      true,
				HasParseError: false,
				Tag:           "returns execute errors list",
			},
		}
		for _, testCase := range vkApiTestCases {
			//nolint:exhaustruct
			customResponse := http.Response{
				Body: io.NopCloser(bytes.NewBufferString(testCase.ResultBody)),
			}

			res, err := responseParser.(*jsonresponseparser.JsonResponseParser).Parse(&customResponse)
			if testCase.HasParseError {
				require.Error(t, err, fmt.Sprintf("%s must have a parse error", testCase.Tag))
			} else {
				require.NoError(t, err, fmt.Sprintf("%s must have not a parse error", testCase.Tag))
			}

			if testCase.HasError {
				require.Error(t, res.Error(), fmt.Sprintf("%s must have response error", testCase.Tag))
			} else {
				require.NoError(t, res.Error(), fmt.Sprintf("%s must have not response error", testCase.Tag))
			}
		}
	})

	t.Run("validation error fields test", func(t *testing.T) {
		t.Parallel()

		res, err := responseParser.Parse(
			//nolint:exhaustruct
			&http.Response{
				Body: io.NopCloser(
					bytes.NewBufferString(`{"error":{"error_msg":"Validation required", "error_code": 17, "redirect_uri":` +
						`"redirect_uri"}}`),
				),
			},
		)
		require.NoError(t, err)

		apiError := res.Error()

		require.Error(t, apiError)

		var errorObject *response.Error
		require.ErrorAs(t, apiError, &errorObject)

		require.Equal(t, errorObject.RedirectUri, "redirect_uri")
	})

	t.Run("captcha error fields test", func(t *testing.T) {
		t.Parallel()

		res, err := responseParser.(*jsonresponseparser.JsonResponseParser).Parse(
			//nolint:exhaustruct
			&http.Response{
				Body: io.NopCloser(
					bytes.NewBufferString(
						`{"error":{"error_msg":"Captcha needed", "error_code":14, "captcha_sid":"1",` +
							`"captcha_img":"https://vk.com/captcha.php?sid=1"}}`,
					),
				),
			})

		require.NoError(t, err)

		apiError := res.Error()
		require.Error(t, apiError)

		var errorObject *response.Error

		require.ErrorAs(t, apiError, &errorObject)
		require.Equal(t, errorObject.CaptchaImg, "https://vk.com/captcha.php?sid=1")
		require.Equal(t, errorObject.CaptchaSid, "1")
	})

	t.Run("error global fields test", func(t *testing.T) {
		t.Parallel()
		res, err := responseParser.(*jsonresponseparser.JsonResponseParser).Parse(
			//nolint:exhaustruct
			&http.Response{
				Body: io.NopCloser(bytes.NewBufferString(
					`{"error":{"error_msg":"Captcha needed", "error_code":14, "captcha_sid":"1",` +
						`"captcha_img":"https://vk.com/captcha.php?sid=1"}}`,
				),
				),
			})
		require.NoError(t, err)

		apiError := res.Error()
		require.Error(t, apiError)

		var errorObject *response.Error

		require.ErrorAs(t, apiError, &errorObject)
		require.Equal(t, errorObject.IntCode(), 14)
		require.Equal(t, errorObject.Error(), "Captcha needed")
	})
}

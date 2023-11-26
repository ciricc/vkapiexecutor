package jsonresponseparser

import (
	"encoding/json"
	"net/http"

	"github.com/buger/jsonparser"
	response "github.com/ciricc/vkapiexecutor/response"
)

// Объект ответа VK API в формате JSON
type JsonResponse struct {
	response.UnknownResponse
}

func NewApiJsonResponse(httpResponse *http.Response) *JsonResponse {
	res := &JsonResponse{
		*response.NewUnknown(httpResponse),
	}
	return res
}

// Валидирует JSON
func (v *JsonResponse) ValidateJson() error {
	t := map[string]interface{}{}
	return json.Unmarshal(v.Body(), &t)
}

// Возвращает информацию об ошибке выполнения метода
func (v *JsonResponse) Error() error {
	body := v.Body()

	var errorObjects [][]byte

	executeErrors, _, _, err := jsonparser.Get(body, "execute_errors")
	isExecuteErrors := false
	if err == nil {
		isExecuteErrors = true
		_, err := jsonparser.ArrayEach(
			executeErrors,
			func(errorObject []byte, dataType jsonparser.ValueType, offset int, err error) {
				errorObjects = append(errorObjects, errorObject)
			},
		)
		if err != nil {
			//nolint:nilerr
			return err
		}
	} else {
		errorObject, _, _, err := jsonparser.Get(body, "error")
		if err != nil {
			//nolint:nilerr
			return nil
		}

		errorObjects = append(errorObjects, errorObject)
	}

	apiErrors := make([]*response.Error, len(errorObjects))

	for i := range errorObjects {
		errorObject := errorObjects[i]

		errorMessage, _ := jsonparser.GetString(errorObject, "error_msg")
		errorMessageIntCode, _ := jsonparser.GetInt(errorObject, "error_code")

		if errorMessage == "" && errorMessageIntCode == 0 {
			return nil
		}

		apiError := response.NewError(errorMessage, int(errorMessageIntCode))

		apiError.RedirectUri, _ = jsonparser.GetString(errorObject, "redirect_uri")
		apiError.CaptchaImg, _ = jsonparser.GetString(errorObject, "captcha_img")
		apiError.CaptchaSid, _ = jsonparser.GetString(errorObject, "captcha_sid")
		apiError.Method, _ = jsonparser.GetString(errorObject, "method")

		apiErrors[i] = apiError
	}

	if len(apiErrors) == 0 {
		return nil
	}

	if len(apiErrors) == 1 && !isExecuteErrors {
		return apiErrors[0]
	}

	return response.NewExecuteErrors(apiErrors)
}

package jsonresponseparser

import (
	"net/http"

	"github.com/buger/jsonparser"
	response "github.com/ciricc/vkapiexecutor/response"
)

// Объект ответа VK API в формате JSON
type JsonResponse struct {
	response.UnknownResponse
}

func NewApiJsonResponse(httpResponse *http.Response) *JsonResponse {
	return &JsonResponse{
		*response.NewUnknown(httpResponse),
	}
}

// Возвращает информацию об ошибке выполнения метода
func (v *JsonResponse) Error() error {
	body := v.Body()

	errorObject, _, _, err := jsonparser.Get(body, "error")
	if err != nil {
		return nil
	}

	errorMessage, _ := jsonparser.GetString(errorObject, "error_msg")
	errorMessageIntCode, _ := jsonparser.GetInt(errorObject, "error_code")

	if errorMessage == "" && errorMessageIntCode == 0 {
		return nil
	}

	return response.NewError(errorMessage, int(errorMessageIntCode))
}

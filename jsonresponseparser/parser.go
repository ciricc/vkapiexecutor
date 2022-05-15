package jsonresponseparser

import (
	"net/http"

	response "github.com/ciricc/vkapiexecutor/response"
)

// Реализует интерфейс парсера (responseparser.ResponseParser) ответа VK API с поддержкой формата JSON
type JsonResponseParser struct{}

// Парсит ответ в формате JSON
func (*JsonResponseParser) Parse(req *http.Response) (response.Response, error) {
	jsonResponse := NewApiJsonResponse(req)
	return jsonResponse, jsonResponse.ValidateJson()
}

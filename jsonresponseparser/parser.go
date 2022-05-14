package jsonresponseparser

import (
	"net/http"

	response "github.com/ciricc/vkapiexecutor/response"
)

// Реализует интерфейс парсера (responseparser.IResponseParser) ответа VK API с поддержкой формата JSON
type JsonResponseParser struct{}

// Парсит ответ в формате JSON
func (*JsonResponseParser) Parse(req *http.Response) (response.IResponse, error) {
	return NewApiJsonResponse(req), nil
}

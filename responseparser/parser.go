package responseparser

import (
	"net/http"

	"github.com/ciricc/vkapiexecutor/response"
)

// Интерфейс парсера тела ответа VK API
type Parser interface {
	Parse(req *http.Response) (response.Response, error) // Парсит тело ответа сервера и возвращает объект релизующий интерфейс стандартнго ответа
}

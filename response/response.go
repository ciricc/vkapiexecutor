package response

import (
	"context"
	"net/http"
)

// Стандартный интерфейс, который понимает executor
type Response interface {
	// Возвращает контекст запроса
	Context() context.Context
	// Возвращает тело ответа в байтах
	Body() []byte
	// Возвращает тело ответа в строковом представлении
	String() string
	// Возвращает информацию об ошибке выполнения запроса
	Error() error
	// Возвращает объект HTTP ответа
	HttpResponse() *http.Response
	// Устанавливате необходимость перевыплнить запрос
	// Используется в основном в хуках
	Renew(renew bool)
	IsRenew() bool
}

package response

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// Стандартный интерфейс, который понимает executor
type Response interface {
	Context() context.Context // Возвращает контекст запроса
	Body() []byte                 // Возвращает тело ответа в байтах
	String() string               // Возвращает тело ответа в строковом представлении
	Error() error                 // Возвращает информацию об ошибке выполнения запроса
	HttpResponse() *http.Response // Возвращает объект HTTP ответа
	Renew(renew bool)
	IsRenew() bool
}

/* Реализация неизвестного формата ответа
   Используется для наследования от него и дальнейшего написания обработки под конкретный формат
*/
type UnknownResponse struct {
	response  *http.Response
	bodyBytes []byte
	renew     bool
}

func NewUnknown(httpResponse *http.Response) *UnknownResponse {
	bodyBytes := bytes.NewBuffer([]byte{})
	io.Copy(bodyBytes, httpResponse.Body)
	return &UnknownResponse{
		response:  httpResponse,
		bodyBytes: bodyBytes.Bytes(),
	}
}

// Возвращает заданный контекст, взятый из запроса request.Request
func (v *UnknownResponse) Context() context.Context {
	return v.response.Request.Context()
}

// Возвращает байты тела ответа сервера
func (v *UnknownResponse) Body() []byte {
	return v.bodyBytes
}

// Возвращает строковое представление тела
func (v *UnknownResponse) String() string {
	return string(v.Body())
}

/* Возвращает информацию об ошибке выполнения метода
   Не возвращает ошибки. связанные с отправкой HTTP запроса
   Метод должен возвращать ошибку типа response.Error
*/
func (v *UnknownResponse) Error() error {
	return nil
}

// Возвращает объект http ответа сервера
func (v *UnknownResponse) HttpResponse() *http.Response {
	return v.response
}

// Просит executor выполнить запрос и создать новый объект Response после выполнения
func (v *UnknownResponse) Renew(renew bool) {
	v.renew = renew
}

// Возвращает информацию о том, нужно ли выполнить запрос снова и пересоздать объект ответа
func (v *UnknownResponse) IsRenew() bool {
	return v.renew
}

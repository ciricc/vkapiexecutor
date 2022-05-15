package request

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// URL путь к VK API
var DefaultBaseRequestUrl = "https://api.vk.com/method/"

// Тип содержимого запроса по умолчанию
const DefaultContentTypeHeaderValue = "application/x-www-form-urlencoded"

type requestContextKey struct{}

// Объект запроса к API ВКонткте
type Request struct {
	method      string      // Метод VK API
	params      *Params     // Параметры запроса
	headers     http.Header // HTTP заголовки запроса. По умолчанию в запросе есть один загловок - Content-Type, его изменить нельзя
	block       bool
	blockReason error
}

// Создает новый API запрос
func New() *Request {
	r := &Request{
		headers: http.Header{},
		params:  NewParams(),
	}
	r.setContentTypeHeader()
	return r
}

// Попытается найти объект запроса в контексте по ключу и, если он есть, вернет его
func FromContext(ctx context.Context) (*Request, error) {
	if val := ctx.Value(requestContextKey{}); val != nil {
		return val.(*Request), nil
	}
	return nil, fmt.Errorf("api request not found in context")
}

// Устанавливает метод VK API и возвращает копию новый запрос
func (v *Request) Method(methodName string) {
	v.method = methodName
}

// Method возвращает текущий метод запроса
func (v *Request) GetMethod() string {
	return v.method
}

// SetParams устанавливает параметры запроса. Глобальные значения при этом не перезаписываются
func (v *Request) Params(params *Params) {
	v.params = params
}

// Params возвращает текущий объект параетров
func (v *Request) GetParams() *Params {
	return v.params
}

// Headers возвращает копию текущих заголовков запроса
func (v *Request) GetHeaders() http.Header {
	if v.headers == nil {
		v.headers = make(http.Header)
	}
	return v.headers
}

/* Устанавливает заголовки, полностью перезаписывает текущие заголовки
Заголовок Content-Type при этом не изменяется, так как Request гарантирует одинаковый формат содержимого
*/
func (v *Request) Headers(headers http.Header) {
	v.headers = headers
	v.setContentTypeHeader()
}

// Сериализирует объект запроса в строку для удобного отображения в логах
func (v *Request) String() string {
	return fmt.Sprintf("url: %q\nmethod: %q\nheaders: %v\nparams: %q", DefaultBaseRequestUrl, v.GetMethod(), v.GetHeaders(), v.GetParams())
}

/* Расширяет текущие заголовки.
   При этом, значение ключа будет перезаписано, если оно уже есть.
*/
func (v *Request) AppendHeaders(headers http.Header) {
	requestHeaders := v.GetHeaders()
	for key, val := range headers {
		requestHeaders[key] = val
	}

	v.setContentTypeHeader()
}

/* Возвращает URL запроса без параметров.
   Метод использует значение переменной request.DefaultBaseRequestUrl
*/
func (v *Request) GetRequestUrl() (*url.URL, error) {
	baseUrl, err := url.Parse(DefaultBaseRequestUrl)
	if err != nil {
		return nil, fmt.Errorf("parse base url variable error: %w", err)
	}
	baseUrl.Path = path.Join(baseUrl.Path, "/", v.method)
	return baseUrl, nil
}

// Возвращает объект запроса для POST метода
func (v *Request) HttpRequestPost() (*http.Request, error) {
	return v.HttpRequest("POST")
}

// Возвращает объект запроса для GET метода
func (v *Request) HttpRequestGet() (*http.Request, error) {
	return v.HttpRequest("GET")
}

// Устанавливает заголовок content-type
func (v *Request) setContentTypeHeader() {
	if v.GetHeaders().Get("content-type") != DefaultContentTypeHeaderValue {
		v.GetHeaders().Set("Content-Type", DefaultContentTypeHeaderValue)
	}
}

/* Возвращает объект http.Request данного API запроса стандартной библиотеки net/http
   По умолчанию все запросы используют заголовок Content-Type: application/x-www-form-urlencoded

   request.HttpRequest("GET")
   request.GttpRequest("POST")
*/
func (v *Request) HttpRequest(httpMethod string) (*http.Request, error) {
	requestUrl, err := v.GetRequestUrl()

	if err != nil {
		return nil, fmt.Errorf("build http request url error: %w", err)
	}

	req := http.Request{
		Method: httpMethod,
		URL:    requestUrl,
		Header: v.GetHeaders(),
	}

	if v.params != nil {
		if httpMethod == "GET" {
			req.URL.RawQuery = v.params.String()
		} else {
			req.Body = io.NopCloser(bytes.NewBuffer([]byte(v.params.String())))
		}
	}

	return &req, nil
}

/* Добавляет текущий запрос в контекст по внутреннему ключу
   ctx := context.Background()
   ctx = req.SetContextValue(ctx)
   request.FromContext(ctx)
*/
func (v *Request) SetContextValue(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestContextKey{}, v)
}

// Блокирует выполнение запроса
func (v *Request) Block(block bool) {
	v.block = block
	if !block {
		v.blockReason = nil
	}
}

// Блокирует выополнение запроса с причиной
func (v *Request) BlockReason(reason error) {
	v.block = true
	v.blockReason = reason
}

// Возвращает информацию о том, нужно либ локировать выполнение запроса
func (v *Request) IsBlock() (bool, error) {
	return v.block, v.blockReason
}

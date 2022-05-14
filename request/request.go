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

type requestContextKey struct{}

// Объект запроса к API ВКонткте
type Request struct {
	method  string      // Метод VK API
	params  *Params     // Параметры запроса
	headers http.Header // HTTP заголовки запроса. По умолчанию в запросе есть один загловок - Content-Type, его изменить нельзя
}

// Создает новый API запрос
func New() *Request {
	return &Request{
		headers: http.Header{},
		params:  NewParams(),
	}
}

// Попытается найти объект запроса в контексте по ключу и, если он есть, вернет его
func FromContext(ctx context.Context) (*Request, error) {
	if val := ctx.Value(requestContextKey{}); val != nil {
		return val.(*Request), nil
	}
	return nil, fmt.Errorf("api request not found in context")
}

// SetMethod устанавливает метод VK API
func (v *Request) Method(methodName string) *Request {
	v.method = methodName
	return v
}

// Method возвращает текущий метод запроса
func (v *Request) GetMethod() string {
	return v.method
}

// SetParams устанавливает параметры запроса. Глобальные значения при этом не перезаписываются
func (v *Request) Params(params *Params) *Request {
	v.params = params
	return v
}

// Params возвращает текущий объект параетров
func (v *Request) GetParams() *Params {
	return v.params
}

// Headers возвращает копию текущих заголовков запроса
func (v *Request) GetHeaders() http.Header {
	return v.headers
}

/* SetHeaders устанавливает заголовки, полностью перезаписывает текущие заголовки
Заголовок Content-Type при этом не изменяется, так как Request гарантирует одинаковый формат содержимого
*/
func (v *Request) Headers(headers http.Header) *Request {
	v.headers = headers
	return v
}

// Сериализирует объект запроса в строку для удобного отображения в логах
func (v *Request) String() string {
	return fmt.Sprintf("url: %q\nmethod: %q\nheaders: %v\nparams: %v", DefaultBaseRequestUrl, v.method, v.headers, v.params)
}

/* Расширяет текущие заголовки.
   При этом, значение ключа будет перезаписано, если оно уже есть.
*/
func (v *Request) AppendHeaders(headers http.Header) *Request {
	for key, val := range headers {
		v.headers[key] = val
	}
	return v
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
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
	}

	if v.params != nil {
		if httpMethod == "GET" {
			req.URL.RawQuery = v.params.String()
		} else {
			req.Body = io.NopCloser(bytes.NewBuffer([]byte(v.params.String())))
		}
	}

	for key, val := range v.headers {
		if _, ok := req.Header[key]; !ok {
			req.Header[key] = val
		}
	}

	return &req, nil
}

/* Добавляет текущий запрос в контекст по внутреннему ключу
   ctx := context.Background()
   ctx = req.SetContextKey(ctx)
   request.FromContext(ctx)
*/
func (v *Request) SetContextKey(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestContextKey{}, v)
}

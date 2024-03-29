package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// URL путь к VK API
var DefaultBaseRequestUrl = "https://api.vk.com/method/"

// Тип содержимого запроса по умолчанию
const DefaultContentTypeHeaderValue = "application/x-www-form-urlencoded"

// Объект запроса к API ВКонткте
type Request struct {
	method          string        // Метод VK API
	params          *Params       // Параметры запроса
	headers         http.Header   // HTTP заголовки запроса. По умолчанию в запросе есть один загловок - Content-Type, его изменить нельзя
	getHttpRequest  *http.Request // HTTP запрос для GET метода
	postHttpRequest *http.Request // HTTP запрос для POST метода
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

// Возвращает и создает, если нет, http запрос для POST метода
func (v *Request) getHttpRequestPost() *http.Request {
	if v.postHttpRequest == nil {
		v.postHttpRequest = &http.Request{
			Method: "POST",
		}
	}
	return v.postHttpRequest
}

// Возвращает и создает, если нет, http запрос для GET метода
func (v *Request) getHttpRequestGet() *http.Request {
	if v.getHttpRequest == nil {
		v.getHttpRequest = &http.Request{
			Method: "GET",
		}
	}
	return v.getHttpRequest
}

// Устанавливает заголовки, полностью перезаписывает текущие заголовки
// Заголовок Content-Type при этом не изменяется, так как Request гарантирует одинаковый формат содержимого
func (v *Request) Headers(headers http.Header) {
	v.headers = headers
	v.setContentTypeHeader()
}

// Сериализирует объект запроса в строку для удобного отображения в логах
func (v *Request) String() string {
	return fmt.Sprintf("url: %q\nmethod: %q\nheaders: %v\nparams: %q", DefaultBaseRequestUrl, v.GetMethod(), v.GetHeaders(), v.GetParams())
}

// Расширяет текущие заголовки.
// При этом, значение ключа будет перезаписано, если оно уже есть.
func (v *Request) AppendHeaders(headers http.Header) {
	requestHeaders := v.GetHeaders()

	for key, val := range headers {
		requestHeaders[key] = val
	}

	v.setContentTypeHeader()
}

/*
Возвращает URL запроса без параметров.

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
	return v.buildHttpRequest("POST")
}

// Возвращает объект запроса для GET метода
func (v *Request) HttpRequestGet() (*http.Request, error) {
	return v.buildHttpRequest("GET")
}

// Устанавливает заголовок content-type
func (v *Request) setContentTypeHeader() {
	if v.GetHeaders().Get("content-type") != DefaultContentTypeHeaderValue {
		v.GetHeaders().Set("Content-Type", DefaultContentTypeHeaderValue)
	}
}

// Возвращает объект http.Request данного API запроса стандартной библиотеки net/http
// По умолчанию все запросы используют заголовок Content-Type: application/x-www-form-urlencoded
// request.HttpRequest("GET")
// request.GttpRequest("POST")
func (v *Request) buildHttpRequest(method string) (*http.Request, error) {
	requestUrl, err := v.GetRequestUrl()

	if err != nil {
		return nil, fmt.Errorf("build http request url error: %w", err)
	}

	var req *http.Request

	if method == "GET" {
		req = v.getHttpRequestGet()
	} else {
		req = v.getHttpRequestPost()
	}

	req.URL = requestUrl
	req.Header = v.GetHeaders()

	if v.params != nil {
		if method == "GET" {
			req.URL.RawQuery = v.params.String()
		} else {
			req.Body = io.NopCloser(strings.NewReader(v.params.String()))
		}
	}

	return req, nil
}

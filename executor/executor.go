package executor

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/ciricc/vkapiexecutor/jsonresponseparser"
	"github.com/ciricc/vkapiexecutor/request"
	"github.com/ciricc/vkapiexecutor/response"
	"github.com/ciricc/vkapiexecutor/responseparser"
)

var DefaultMaxRequestTries = 50
var DefaultResponseParser responseparser.Parser = &jsonresponseparser.JsonResponseParser{}

type ApiResponseHandlerNext func(res response.Response) error
type ApiResponseHandler func(next ApiResponseHandlerNext, res response.Response) error
type HttpResponseHandlerNext func(res *http.Response) error
type HttpResponseHandler func(next HttpResponseHandlerNext, res *http.Response) error

// Отвечает за выполнение API запросов ВКонтакте
type Executor struct {
	HttpClient         *http.Client          // HTTP клиент для отправки запросов. Вы можете задать свой клиент, настроив, например, прокси или KeepAlive соединение
	ResponseParser     responseparser.Parser // Парсер ответа ВКонтакте. Можно переназначить для парсинга других форматов
	apiResponseHandle  ApiResponseHandler    // Последний добавленный обработчик API ответа
	httpResponseHandle HttpResponseHandler   // Последний добавленный обработчик HTTP ответа
	MaxRequestTries    int
}

type requestTryContextKey struct{} // Ключ счетчика попыток запроса в контексте
type requestContextKey struct{}    // Ключ запроса в контексте

var requestContextKeyVal = requestContextKey{}
var requestTryContextKeyVal = requestTryContextKey{}

func New() *Executor {
	return &Executor{
		HttpClient:        http.DefaultClient,
		ResponseParser:    DefaultResponseParser,
		apiResponseHandle: func(next ApiResponseHandlerNext, res response.Response) error { return nil },
		MaxRequestTries:   DefaultMaxRequestTries,
		httpResponseHandle: func(next HttpResponseHandlerNext, res *http.Response) error {
			return nil
		},
	}
}

// Устанавливает middleware для обработки ответа VK API
func (v *Executor) HandleApiResponse(handler ApiResponseHandler) {
	nextHandler := v.apiResponseHandle
	v.apiResponseHandle = func(next ApiResponseHandlerNext, res response.Response) error {
		return handler(func(res response.Response) error {
			return nextHandler(nil, res)
		}, res)
	}
}

// Устанавливает обработчик ответов сервера
func (v *Executor) HandleHttpResponse(handler HttpResponseHandler) {
	nextHandler := v.httpResponseHandle
	v.httpResponseHandle = func(next HttpResponseHandlerNext, res *http.Response) error {
		return handler(func(res *http.Response) error {
			return nextHandler(nil, res)
		}, res)
	}
}

// Отчищает очередь из middleware API ответов
func (v *Executor) ResetApiResponseHandlers() {
	v.apiResponseHandle = func(next ApiResponseHandlerNext, res response.Response) error { return nil }
}

// Отчищает очередь из обработчиков HTTP овтетов
func (v *Executor) ResetHttpResponseHandlers() {
	v.httpResponseHandle = func(next HttpResponseHandlerNext, res *http.Response) error { return nil }
}

// Выполняет запрос к VK API
// Используйте executor.DoRequestCtx(), если есть задача контролировать таймаут и контекст запроса
func (v *Executor) DoRequest(req *request.Request) (response.Response, error) {
	return v.DoRequestCtx(context.Background(), req)
}

// Выполняет запрос к VK API.
func (v *Executor) DoRequestCtx(ctx context.Context, req *request.Request) (response.Response, error) {
	return v.DoRequestCtxParser(ctx, req, v.ResponseParser)
}

// Выполняет запрос к VK API.
// Используйте ctx для передачи значений в middleware или для создания таймаутов на выполнение запроса
// Вы можете задать свой собственный парсер ответа. Например, ВКонтакте поддерживает формат messagepack (users.get.msgpack)
// Возвращает ответ VK API. В случае, если возникла ошибка выоплнения HTTP запроса, то будет response.Response == nil.
// Если возникла ошибка при вызове метода API, вернется полный ответ сервера и информация об ошибке типа response.Error
func (v *Executor) DoRequestCtxParser(ctx context.Context, req *request.Request, parser responseparser.Parser) (response.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("input request empty")
	}

	if parser == nil {
		return nil, fmt.Errorf("response parser is nil")
	}

	var requestTry = getRequestTryPtr(ctx)
	if requestTry == nil {
		reqTry := int32(0)
		requestTry = &reqTry

		// Устанавливаем в контекст все необходимые данные запроса
		// Во избежание ошибок - контекстом управляет только Executor,
		// поэтому именно он и может получить из контекста запрос и счетчик выполнений
		ctx = context.WithValue(ctx, requestTryContextKeyVal, requestTry)
		ctx = context.WithValue(ctx, requestContextKeyVal, req)
	}

	if atomic.LoadInt32(requestTry) >= int32(v.MaxRequestTries) {
		return nil, fmt.Errorf("max request calls exceeded: %d", v.MaxRequestTries)
	}

	httpReq, err := req.HttpRequestPost()
	if err != nil {
		return nil, err
	}

	httpReq = httpReq.WithContext(ctx)

	res, err := v.HttpClient.Do(httpReq)
	addRequestTry(requestTry)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	err = v.httpResponseHandle(nil, res)
	if err != nil {
		return nil, err
	}

	apiResponse, err := parser.Parse(res)
	res.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}

	err = v.apiResponseHandle(nil, apiResponse)
	if err != nil {
		return nil, err
	}

	if apiResponse.IsRenew() {
		return v.DoRequestCtxParser(ctx, req, parser)
	}

	return apiResponse, apiResponse.Error()
}

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

type ApiRequestHandlerNext func(ctx context.Context, req *request.Request) error
type ApiRequestHandler func(next ApiRequestHandlerNext, ctx context.Context, req *request.Request) error

type HttpRequestHandlerNext func(req *http.Request) error
type HttpRequestHandler func(next HttpRequestHandlerNext, req *http.Request) error

type ApiResponseHandlerNext func(res response.Response) error
type ApiResponseHandler func(next ApiResponseHandlerNext, res response.Response) error

// Отвечает за выполнение API запросов ВКонтакте
type Executor struct {
	HttpClient        *http.Client          // HTTP клиент для отправки запросов. Вы можете задать свой клиент, настроив, например, прокси или KeepAlive соединение
	ResponseParser    responseparser.Parser // Парсер ответа ВКонтакте. Можно переназначить для парсинга других форматов
	apiRequestHandle  ApiRequestHandler     // Последний добавленный обработчик API запроса
	httpRequestHandle HttpRequestHandler    // Последний добавленный обработчик HTTP запроса
	apiResponseHandle ApiResponseHandler    // Последлний добавленный обработчик API ответа
	MaxRequestTries   int
}

type requestTryContextKey struct{} // Ключ счетчика попыток запроса в контексте
type requestContextKey struct{}    // Ключ запроса в контексте

func New() *Executor {
	return &Executor{
		HttpClient:        http.DefaultClient,
		ResponseParser:    &jsonresponseparser.JsonResponseParser{},
		apiRequestHandle:  func(_ ApiRequestHandlerNext, _ context.Context, _ *request.Request) error { return nil },
		httpRequestHandle: func(_ HttpRequestHandlerNext, _ *http.Request) error { return nil },
		apiResponseHandle: func(next ApiResponseHandlerNext, res response.Response) error { return nil },
		MaxRequestTries:   DefaultMaxRequestTries,
	}
}

/* Устанавливает middleware для обработки запроса VK API
   Обязательно возвращайте вызов функции next(), если нет цели прервать выполнение всех предыдущих middlware
*/
func (v *Executor) HandleApiRequest(handler ApiRequestHandler) {
	nextHandler := v.apiRequestHandle
	v.apiRequestHandle = func(next ApiRequestHandlerNext, ctx context.Context, req *request.Request) error {
		return handler(func(ctx context.Context, req *request.Request) error {
			return nextHandler(nil, ctx, req)
		}, ctx, req)
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

// Отчищает очередь из middleware VK API запросов
func (v *Executor) ResetApiRequestHandlers() {
	v.apiRequestHandle = func(_ ApiRequestHandlerNext, _ context.Context, _ *request.Request) error { return nil }
}

// Отчищает очередь из middleware HTTP запросов
func (v *Executor) ResetHttpRequestHandlers() {
	v.httpRequestHandle = func(next HttpRequestHandlerNext, req *http.Request) error { return nil }
}

// Отчищает очередь из middleware API ответов
func (v *Executor) ResetApiResponseHandlers() {
	v.apiResponseHandle = func(next ApiResponseHandlerNext, res response.Response) error { return nil }
}

/* Устанавливает middleware для обработки HTTP запроса
   Обязательно возвращайте вызов функции next(), если нет цели прервать выполнение всех предыдущих middlware
*/
func (v *Executor) HandleHttpRequest(handler HttpRequestHandler) {
	nextHandler := v.httpRequestHandle
	v.httpRequestHandle = func(next HttpRequestHandlerNext, req *http.Request) error {
		return handler(func(req *http.Request) error {
			return nextHandler(nil, req)
		}, req)
	}
}

/* Выполняет запрос к VK API
   Используйте executor.DoRequestCtx(), если есть задача контролировать таймаут и контекст запроса
*/
func (v *Executor) DoRequest(req *request.Request) (response.Response, error) {
	return v.DoRequestCtx(context.Background(), req)
}

// Выполняет запрос к VK API.
func (v *Executor) DoRequestCtx(ctx context.Context, req *request.Request) (response.Response, error) {
	return v.DoRequestCtxParser(ctx, req, v.ResponseParser)
}

/*
   Выполняет запрос к VK API.
   Используйте ctx для передачи значений в middleware или для создания таймаутов на выполнение запроса
   Вы можете задать свой собственный парсер ответа. Например, ВКонтакте поддерживает формат messagepack (users.get.msgpack)

   Возвращает ответ VK API. В случае, если возникла ошибка выоплнения HTTP запроса, то будет response.Response == nil.
   Если возникла ошибка при вызове метода API, вернется полный ответ сервера и информация об ошибке типа response.Error
*/
func (v *Executor) DoRequestCtxParser(ctx context.Context, req *request.Request, parser responseparser.Parser) (response.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("input request empty")
	}

	if parser == nil {
		return nil, fmt.Errorf("response parser is nil")
	}

	err := v.apiRequestHandle(nil, ctx, req)
	if err != nil {
		return nil, err
	}

	err = getCantDoRequestReason(req)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %w", err)
	}

	var requestTry = getRequestTryPtr(ctx)

	if requestTry == nil {
		reqTry := int32(0)
		requestTry = &reqTry

		/* Устанавливаем в контекст все необходимые данные запроса
		   Во избежание ошибок - контекстом управляет только Executor,
		   поэтому именно он и может получить из контекста запрос и счетчик выполнений
		*/
		ctx = context.WithValue(ctx, requestTryContextKey{}, requestTry)
		ctx = context.WithValue(ctx, requestContextKey{}, req)
	}

	if atomic.LoadInt32(requestTry) >= int32(v.MaxRequestTries) {
		return nil, fmt.Errorf("max request calls exceeded: %d", v.MaxRequestTries)
	}

	httpReq, err := req.HttpRequestPost()
	if err != nil {
		return nil, err
	}

	httpReq = httpReq.WithContext(ctx)

	err = v.httpRequestHandle(nil, httpReq)
	if err != nil {
		return nil, err
	}

	addRequestTry(requestTry)
	res, err := v.HttpClient.Do(httpReq)

	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	apiResponse, err := parser.Parse(res)
	defer res.Body.Close()

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

package executor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ciricc/vkapiexecutor/jsonresponseparser"
	"github.com/ciricc/vkapiexecutor/request"
	"github.com/ciricc/vkapiexecutor/response"
	"github.com/ciricc/vkapiexecutor/responseparser"
)

type ApiRequestHandlerNext func(ctx context.Context, req *request.Request) error
type ApiRequestHandler func(next ApiRequestHandlerNext, ctx context.Context, req *request.Request) error

type HttpRequestHandlerNext func(req *http.Request) error
type HttpRequestHandler func(next HttpRequestHandlerNext, req *http.Request) error

// Отвечает за выполнение API запросов ВКонтакте
type Executor struct {
	HttpClient        *http.Client          // HTTP клиент для отправки запросов. Вы можете задать свой клиент, настроив, например, прокси или KeepAlive соединение
	ResponseParser    responseparser.Parser // Парсер ответа ВКонтакте. Можно переназначить для парсинга других форматов
	apiRequestHandle  ApiRequestHandler     // Последний добавленный обработчик API запроса
	httpRequestHandle HttpRequestHandler    // Последний добавленный обработчик HTTP запроса
}

func New() *Executor {
	httpClient := http.Client{}
	return &Executor{
		HttpClient:        &httpClient,
		ResponseParser:    &jsonresponseparser.JsonResponseParser{},
		apiRequestHandle:  func(_ ApiRequestHandlerNext, _ context.Context, _ *request.Request) error { return nil },
		httpRequestHandle: func(_ HttpRequestHandlerNext, _ *http.Request) error { return nil },
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

// Отчищает очередь из middleware VK API запросов
func (v *Executor) ResetApiRequestHandlers() {
	v.apiRequestHandle = func(_ ApiRequestHandlerNext, _ context.Context, _ *request.Request) error { return nil }
}

// Отчищает очередь из middleware HTTP запросов
func (v *Executor) ResetHttpRequestHandlers() {
	v.httpRequestHandle = func(next HttpRequestHandlerNext, req *http.Request) error { return nil }
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

// Возвращает причину, по которой нельзя выполнить запрос
func (v *Executor) getCantDoRequestReason(req *request.Request) error {
	if blocked, reason := req.IsBlock(); blocked {
		if reason != nil {
			return fmt.Errorf("request blocked: %w", reason)
		}
		return fmt.Errorf("request blocked")
	}
	return nil
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

	if req == nil {
		return nil, fmt.Errorf("after middleware chanining input request is empty")
	}

	err = v.getCantDoRequestReason(req)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %w", err)
	}

	reqCtx := req.SetContextValue(ctx)
	httpReq, err := req.HttpRequestPost()

	if err != nil {
		return nil, err
	}

	err = v.httpRequestHandle(nil, httpReq)
	if err != nil {
		return nil, err
	}

	if httpReq == nil {
		return nil, fmt.Errorf("after middlewate chaining http request is empty")
	}

	if ctx, _ := request.FromContext(reqCtx); ctx == nil {
		return nil, fmt.Errorf("after middleware chaining api request context value is nil")
	}

	httpReq = httpReq.WithContext(reqCtx)
	res, err := v.HttpClient.Do(httpReq)

	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	apiResponse, err := parser.Parse(res)
	defer res.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}

	return apiResponse, apiResponse.Error()
}

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

var (
	DefaultResponseParser responseparser.Parser = &jsonresponseparser.JsonResponseParser{}
)

type ApiResponseNextHook func(res response.Response) error

type ApiResponseHook func(next ApiResponseNextHook, res response.Response) error

type HttpResponseNextHook func(res *http.Response) error

type HttpResponseHook func(next HttpResponseNextHook, res *http.Response) error

// Отвечает за выполнение API запросов ВКонтакте
type Executor struct {
	// HTTP клиент для отправки запросов.
	// Вы можете задать свой клиент, настроив, например, прокси или KeepAlive соединение
	HttpClient *http.Client
	// Парсер ответа ВКонтакте. Можно переназначить для парсинга других форматов
	ResponseParser responseparser.Parser

	// Последний добавленный обработчик API ответа
	apiResponseHook ApiResponseHook
	// Последний добавленный обработчик HTTP ответа
	httpResponseHook HttpResponseHook
}

func New() *Executor {
	return &Executor{
		HttpClient:      http.DefaultClient,
		ResponseParser:  DefaultResponseParser,
		apiResponseHook: func(next ApiResponseNextHook, res response.Response) error { return nil },
		httpResponseHook: func(next HttpResponseNextHook, res *http.Response) error {
			return nil
		},
	}
}

// Устанавливает хук для обработки ответа VK API
func (v *Executor) ApiResponseHook(hook ApiResponseHook) {
	nextHook := v.apiResponseHook
	v.apiResponseHook = func(next ApiResponseNextHook, res response.Response) error {
		return hook(func(res response.Response) error {
			return nextHook(nil, res)
		}, res)
	}
}

// Устанавливает обработчик ответов сервера
func (v *Executor) HttpResponseHook(hook HttpResponseHook) {
	nextHook := v.httpResponseHook
	v.httpResponseHook = func(next HttpResponseNextHook, res *http.Response) error {
		return hook(func(res *http.Response) error {
			return nextHook(nil, res)
		}, res)
	}
}

// Отчищает очередь из middleware API ответов
func (v *Executor) ResetApiResponseHandlers() {
	v.apiResponseHook = func(next ApiResponseNextHook, res response.Response) error { return nil }
}

// Отчищает очередь из обработчиков HTTP овтетов
func (v *Executor) ResetHttpResponseHandlers() {
	v.httpResponseHook = func(next HttpResponseNextHook, res *http.Response) error { return nil }
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

	ctx = context.WithValue(ctx, requestContextKeyVal, req)

	httpReq, err := req.HttpRequestPost()
	if err != nil {
		return nil, err
	}

	httpReq = httpReq.WithContext(ctx)

	res, err := v.HttpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	err = v.httpResponseHook(nil, res)
	if err != nil {
		return nil, err
	}

	apiResponse, err := parser.Parse(res)
	res.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}

	err = v.apiResponseHook(nil, apiResponse)
	if err != nil {
		return nil, err
	}

	if apiResponse.IsRenew() {
		return v.DoRequestCtxParser(ctx, req, parser)
	}

	return apiResponse, apiResponse.Error()
}

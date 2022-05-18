package executor

import (
	"context"
	"sync/atomic"

	"github.com/ciricc/vkapiexecutor/request"
)

// Возвращает объект запроса из контекста
func GetRequest(ctx context.Context) *request.Request {
	if ctx != nil {
		if req, ok := ctx.Value(requestContextKeyVal).(*request.Request); ok {
			return req
		}
	}
	return nil
}

// Возвращает счетчик попыток отправки запроса (уже после всех middleware)
func GetRequestTry(ctx context.Context) int32 {
	reqTry := getRequestTryPtr(ctx)
	if reqTry == nil {
		return 0
	}
	return atomic.LoadInt32(reqTry)
}

// Возвращает из контекста счетчик количества выполнений запроса
func getRequestTryPtr(ctx context.Context) *int32 {

	if ctx == nil {
		return nil
	}

	val := ctx.Value(requestTryContextKey{})
	if val == nil {
		return nil
	}
	return val.(*int32)
}

// Увеличивает счетчик выполнений запроса в контексте
func addRequestTry(try *int32) {
	atomic.AddInt32(try, 1)
}

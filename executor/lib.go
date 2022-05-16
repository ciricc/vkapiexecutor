package executor

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/ciricc/vkapiexecutor/request"
)

// Возвращает объект запроса из контекста
func GetRequest(ctx context.Context) *request.Request {
	req := ctx.Value(requestContextKey{})
	if req == nil {
		return nil
	}
	return req.(*request.Request)
}

// Возвращает причину, по которой нельзя выполнить запрос
func getCantDoRequestReason(req *request.Request) error {
	if blocked, reason := req.IsBlock(); blocked {
		if reason != nil {
			return fmt.Errorf("request blocked: %w", reason)
		}
		return fmt.Errorf("request blocked")
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

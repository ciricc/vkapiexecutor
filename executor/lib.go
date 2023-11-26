package executor

import (
	"context"

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

package executor

// Ключ счетчика попыток запроса в контексте
type requestTryContextKey struct{}

// Ключ запроса в контексте
type requestContextKey struct{}

var (
	requestContextKeyVal    = requestContextKey{}
	requestTryContextKeyVal = requestTryContextKey{}
)

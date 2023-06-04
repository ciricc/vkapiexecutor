package limiter

import (
	"net/http"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
)

// LimiterTripper type
type LimiterTripper http.RoundTripper

// Объект, хранящий в себе все лимитеры запросов на конкретный токен.
// Чтобы не засорять память, в случае, если у вас много (сотни или тысячи) различных токенов,
// в этом пакете используется кеширование лимитеров. Лимитеры токенов регулярно удаляются из памяти.
type tripper struct {
	store   LimiterStore
	Tripper http.RoundTripper
}

// Возвращает новый лимитер.
// rps - Количество запросов в секунду на один токен
// limiterExpiration - Время жизни одного лимитера в памяти.
// После истечения этого времени, он будет пересоздан или автоматически удален.
// cacheCleanupInterval - Интервал, указывающий, как часто нужно отчищать лимитеры, срок жизни которых истек
func New(rps int, limiterExpiration, cacheCleanupInterval time.Duration) LimiterTripper {
	return NewWithStore(
		NewLimiterStoreTtlCache(rps, limiterExpiration, cacheCleanupInterval),
	)
}

// NewWithStore используется для переназначения хранилища лимитеров
// и более точного определения алгоритма лимитеров
func NewWithStore(store LimiterStore) LimiterTripper {
	return &tripper{
		store:   store,
		Tripper: http.DefaultTransport,
	}
}

// RoundTrip возвращает функцию middleware для пакета executor
func (v *tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	apiReq := executor.GetRequest(req.Context())
	if apiReq == nil {
		return nil, ErrRequestEmpty
	}

	limiter, err := v.store.GetLimiter(apiReq)
	if err != nil {
		return nil, err
	}

	err = limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	return v.Tripper.RoundTrip(req)
}

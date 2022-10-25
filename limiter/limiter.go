package limiter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

// Объект, хранящий в себе все лимитеры запросов на конкретный токен.
// Чтобы не засорять память, в случае, если у вас много (сотни или тысячи) различных токенов,
// в этом пакете используется кеширование лимитеров. Лимитеры токенов регулярно удаляются из памяти.
type Limiter struct {
	rateLimit     rate.Limit   // Рейт лимит
	limitersCache *cache.Cache // Кеш лимитеров для каждого токена
	Tripper       http.RoundTripper
}

// Возвращает новый лимитер.
// rps - Количество запросов в секунду на один токен
// limiterExpiration - Время жизни одного лимитера в памяти.
// После истечения этого времени, он будет пересоздан или автоматически удален.
// cacheCleanupInterval - Интервал, указывающий, как часто нужно отчищать лимитеры, срок жизни которых истек
func New(rps int, limiterExpiration time.Duration, cacheCleanupInterval time.Duration) *Limiter {
	return &Limiter{
		rateLimit:     rate.Every((1 * time.Second) / time.Duration(rps)),
		limitersCache: cache.New(limiterExpiration, cacheCleanupInterval),
		Tripper:       http.DefaultTransport,
	}
}

// Возвращает функцию middleware для пакета executor
func (v *Limiter) RoundTrip(req *http.Request) (*http.Response, error) {
	apiRequest := executor.GetRequest(req.Context())
	if apiRequest == nil {
		return nil, fmt.Errorf("not found request in context")
	}

	params := apiRequest.GetParams()

	if params != nil {
		token := params.GetAccessToken()
		if token == "" {
			token = params.GetAnonymousToken()
		}

		if token != "" {
			var limiter *rate.Limiter

			if savedLimiter, ok := v.limitersCache.Get(token); ok {
				limiter = savedLimiter.(*rate.Limiter)
			} else {
				limiter = rate.NewLimiter(v.rateLimit, 1)
				v.limitersCache.Set(token, limiter, cache.DefaultExpiration)
			}

			limiter.Wait(req.Context())
		}
	}

	return v.Tripper.RoundTrip(req)
}

package limiter

import (
	"time"

	"github.com/ciricc/vkapiexecutor/request"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

// LimiterStore это интерфейс для хранения всех лимитеров
// Вы можете реализовать свой собственный store, чтобы точнее контролировать
// лимитирование запросов, а также используемую оперативную память
type LimiterStore interface {
	// GetLimiter возвращает WaitLimiter для дальнейшего вызова его в запросе Tripper'а
	// На вход получет запрос от executor'а, по которому можно определить, например
	// токен доступа, по которому делается запрос, тем самым, создав лимитер под каждый
	// токен отдельно
	GetLimiter(*request.Request) (WaitLimiter, error)
}

type limiterStoreTtlCache struct {
	limitersCache *cache.Cache
	rateLimit     rate.Limit
}

func (c *limiterStoreTtlCache) GetLimiter(
	req *request.Request,
) (WaitLimiter, error) {

	params := req.GetParams()
	if params == nil {
		return DefaultLimiter, nil
	}

	token := params.GetAccessToken()
	if token == "" {
		token = params.GetAnonymousToken()
	}

	if token == "" {
		return DefaultLimiter, nil
	}

	var limiter *rate.Limiter
	if savedLimiter, ok := c.limitersCache.Get(token); ok {
		return savedLimiter.(*rate.Limiter), nil
	} else {
		limiter = rate.NewLimiter(c.rateLimit, 1)
		c.limitersCache.Set(token, limiter, cache.DefaultExpiration)
		return limiter, nil
	}
}

// NewLimiterStoreTtlCache returns limiter's store with ttl cache storage
func NewLimiterStoreTtlCache(
	rps int,
	limiterExpiration,
	cleanupCacheInterval time.Duration,
) LimiterStore {
	return &limiterStoreTtlCache{
		rateLimit:     rate.Every((1 * time.Second) / time.Duration(rps)),
		limitersCache: cache.New(limiterExpiration, cleanupCacheInterval),
	}
}

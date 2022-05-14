Простой модуль для отправки и обработки запросов к VK API, включающий в себя минимум необходимых возможностей.

```go
import (
	"fmt"
	"log"
	"os"
	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/request"
)

func main() {
	token := os.Getenv("TOKEN")
	exec := executor.New()

	requestParams := request.NewParams()
	requestParams.AccessToken = token

	defaultRequest := request.New().Params(requestParams)

	usersGetRequest := copyReq(defaultRequest).Method("users.get")
	statusGetRequest := copyReq(defaultRequest).Method("status.get")

	usersGetResponse, err := exec.DoRequest(usersGetRequest)

	if err != nil {
		panic(fmt.Errorf("users get error: %w", err))
	}

	statusGetResponse, err := exec.DoRequest(statusGetRequest)

	if err != nil {
		panic(fmt.Errorf("status get error: %w", err))
	}

	log.Println("users.get: ", usersGetResponse)
	log.Println("status: ", statusGetResponse)
}

func copyReq(origin *request.Request) *request.Request {
	r := *origin
	return &r
}
```

## Лимитирование запросов
Избегайте ошибок работы из-за превышения количества запросов в секунду.

```go
import (
	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/limiter"
	"github.com/ciricc/vkapiexecutor/request"
)

// 2 запроса в секунду
reqLimiter := limiter.New(2, time.Hour, time.Hour)

exec := executor.New()
exec.HandleApiRequest(reqLimiter.Handle())

params := request.NewParams()
params.AccessToken = token

req := request.New().Method("users.get").Params(params)

wg := sync.WaitGroup{}

for i := 0; i < 100; i++ {
	wg.Add(1)
	go (func() {
		defer wg.Done()
		res, err := exec.DoRequest(req)
		if err != nil {
			panic(fmt.Errorf("request error: %w", err))
		} else {
			log.Println("result", res)
		}
	})()
}

wg.Wait()
```
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
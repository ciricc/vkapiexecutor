package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/limiter"
	"github.com/ciricc/vkapiexecutor/request"
)

func main() {
	token := os.Getenv("TOKEN")

	reqLimiter := limiter.New(2, time.Hour, time.Hour)

	exec := executor.New()
	exec.HandleApiRequest(reqLimiter.Handle())

	params := request.NewParams()
	params.AccessToken(token)

	req := request.New()

	req.Method("users.get")
	req.Params(params)

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
}

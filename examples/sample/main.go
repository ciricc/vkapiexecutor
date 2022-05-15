package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/request"
)

func main() {
	token := os.Getenv("TOKEN")
	exec := executor.New()

	requestParams := request.NewParams()
	requestParams.AccessToken(token)

	defaultRequest := request.New()
	defaultRequest.Params(requestParams)

	usersGetRequest := request.New()
	usersGetRequest.Method("users.get")
	copyRequestParams(usersGetRequest, defaultRequest)
	usersGetRequest.GetParams().Set("user_ids", "1")

	statusGetRequest := request.New()
	statusGetRequest.Method("status.get")
	copyRequestParams(statusGetRequest, defaultRequest)

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

func copyRequestParams(dst *request.Request, source *request.Request) error {
	urlValues, err := url.ParseQuery(source.GetParams().String())
	if err != nil {
		return err
	}
	newParams := request.NewParamsFromUrl(urlValues)
	dst.Params(newParams)
	return nil
}

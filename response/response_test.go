package response_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ciricc/vkapiexecutor/request"
	"github.com/ciricc/vkapiexecutor/response"
)

func TestResponseWithoutContext(t *testing.T) {
	req := request.New()
	httpReq, err := req.HttpRequestGet()

	if err != nil {
		t.Error(err)
	}

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Error(err)
	}

	defer httpRes.Body.Close()

	t.Run("error handle on unknown response is nil", func(t *testing.T) {
		res := response.NewUnknown(httpRes)
		if res.Error() != nil {
			t.Errorf("error is not nil: %s", res.Error())
		}
	})

	t.Run("context will return event if is not used in request", func(t *testing.T) {
		res := response.NewUnknown(httpRes)
		if res.Context() == nil {
			t.Errorf("returned context")
		}
	})

	t.Run("same body with http response body", func(t *testing.T) {
		res := response.NewUnknown(httpRes)
		resBytes, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			t.Error(err)
		}

		if string(res.Body()) != string(resBytes) {
			t.Errorf("different response body:\nexpected: %q\nreal: %q", string(resBytes), string(res.Body()))
		}

		if string(resBytes) != res.String() {
			t.Errorf("response string diffrent with expected value:\nexptected: %q\nreal: %q", string(resBytes), res.String())
		}
	})

	t.Run("return correct http response", func(t *testing.T) {
		res := response.NewUnknown(httpRes)
		if res.HttpResponse() != httpRes {
			t.Errorf("different http response")
		}
	})
}

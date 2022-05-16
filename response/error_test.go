package response_test

import (
	"testing"

	"github.com/ciricc/vkapiexecutor/response"
)

func TestError(t *testing.T) {
	t.Run("values error", func(t *testing.T) {
		err := response.NewError("1", 1)

		if err.IntCode() != 1 {
			t.Errorf("expected int code: %d, real : %d", 1, err.IntCode())
		}

		if err.Error() != "1" {
			t.Errorf("expected message: %q, real: %q", "1", err.Error())
		}
	})
}

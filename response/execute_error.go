package response

import "fmt"

type ExecuteErrors struct {
	errors []*Error
}

func (e *ExecuteErrors) Error() string {
	return fmt.Sprintf("execute errors: %v", e.errors)
}

func (e *ExecuteErrors) Errors() []*Error {
	return e.errors
}

func NewExecuteErrors(errors []*Error) *ExecuteErrors {
	return &ExecuteErrors{
		errors: errors,
	}
}

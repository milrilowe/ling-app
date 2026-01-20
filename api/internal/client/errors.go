package client

import "fmt"

// MLServiceError represents a structured error from the ML service
type MLServiceError struct {
	Code      string
	Message   string
	Retryable bool
}

func (e *MLServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

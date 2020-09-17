package nano

import (
	"fmt"
)

type NodeError struct {
	Message *string `json:"error"`
}

func (e *NodeError) Error() string {
	return *e.Message
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTPError(status=%d, body=%q)", e.StatusCode, e.Body)
}

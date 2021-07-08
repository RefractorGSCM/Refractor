package domain

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

var (
	ErrConflict = errors.New("conflict")  // action cannot be performed
	ErrInternal = errors.New("internal")  // internal server error
	ErrInvalid  = errors.New("invalid")   // validation failed
	ErrNotFound = errors.New("not found") // entity does not exist
	ErrMarshal  = errors.New("marshal")
)

type ClientError interface {
	Error() string

	// ResponseBody returns a response body
	ResponseBody() ([]byte, error)

	// ResponseHeaders returns http status code and headers
	ResponseHeaders() (int, map[string]string)
}

// HTTPError implements the ClientError interface.
type HTTPError struct {
	Cause            error             `json:"-"`
	Message          string            `json:"message"`
	ValidationErrors map[string]string `json:"errors,omitempty"`
	Status           int               `json:"-"`
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s : %s", e.Message, e.Cause.Error())
}

// ResponseBody returns JSON response body.
func (e *HTTPError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("could not parse response body. Error: %v", err)
	}
	return body, nil
}

// ResponseHeaders returns http status code and headers.
func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

func NewHTTPError(err error, status int, detail string) error {
	return &HTTPError{
		Cause:   err,
		Message: detail,
		Status:  status,
	}
}

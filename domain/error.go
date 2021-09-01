/*
 * This file is part of Refractor.
 *
 * Refractor is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package domain

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

var (
	ErrConflict = errors.New("conflict")  // action cannot be performed
	ErrInvalid  = errors.New("invalid")   // validation failed
	ErrNotFound = errors.New("not found") // entity does not exist
	ErrNoArgs   = errors.New("no args")   // no arguments provided
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
	Success          bool        `json:"success"`
	Cause            error       `json:"-"`
	Message          string      `json:"message"`
	ValidationErrors interface{} `json:"errors,omitempty"`
	Status           int         `json:"-"`
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
		Success: false,
		Cause:   err,
		Message: detail,
		Status:  status,
	}
}

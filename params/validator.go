package params

import (
	"Refractor/domain"
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"
)

// ValidateStruct is a wrapper function which calls the wrapError function around validation.ValidateStruct.
func ValidateStruct(structPtr interface{}, fields ...*validation.FieldRules) error {
	return wrapError(validation.ValidateStruct(structPtr, fields...))
}

func wrapError(err error) error {
	if err == nil {
		return err
	}

	httpError := &domain.HTTPError{
		Cause:   domain.ErrInvalid,
		Message: "Input errors exist",
		Status:  http.StatusBadRequest,
	}

	data, err := json.Marshal(err)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &httpError.ValidationErrors)
	if err != nil {
		return err
	}

	return httpError
}

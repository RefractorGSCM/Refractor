package params

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type CreateUserParams struct {
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
}

func (body CreateUserParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Username, validation.Required, validation.Length(1, 20)),
		validation.Field(&body.Email, validation.Required, is.Email),
	)
}

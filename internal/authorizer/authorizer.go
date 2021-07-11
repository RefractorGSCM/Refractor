package authorizer

import (
	"Refractor/domain"
)

type authorizer struct {
}

func NewAuthorizer(idkyet interface{}) domain.Authorizer {
	return &authorizer{}
}

func (a *authorizer) HasPermission(userID, domain, object, action string) (bool, error) {
	// TODO: Figure out how to make this useable
	return false, nil
}

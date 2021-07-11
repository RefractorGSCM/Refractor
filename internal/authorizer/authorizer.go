package authorizer

import (
	"Refractor/domain"
	"github.com/casbin/casbin/v2"
)

type authorizer struct {
	enforcer *casbin.Enforcer
}

func NewAuthorizer(enforcer *casbin.Enforcer) domain.Authorizer {
	return &authorizer{
		enforcer: enforcer,
	}
}

func (a *authorizer) HasPermission(userID, domain, object, action string) (bool, error) {
	return a.enforcer.Enforce(userID, domain, object, action)
}

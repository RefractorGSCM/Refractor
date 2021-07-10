package authorizer

import "github.com/casbin/casbin/v2"

type authorizer struct {
	enforcer *casbin.Enforcer
}

func (a *authorizer) HasPermission(userID, domain, object, action string) (bool, error) {
	return a.enforcer.Enforce(userID, domain, object, action)
}

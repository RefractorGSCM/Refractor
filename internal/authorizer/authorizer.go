package authorizer

import (
	"Refractor/domain"
	"fmt"
	"math/big"
)

type authorizer struct {
}

func NewAuthorizer() domain.Authorizer {
	return &authorizer{}
}

func (a *authorizer) HasPermission(scope domain.AuthScope, userID string, requiredFlags []*big.Int) (bool, error) {
	switch scope.Type {
	case domain.AuthObjRefractor:
		return a.hasPermissionRefractor(userID, requiredFlags)

	case domain.AuthObjServer:
		serverID, ok := scope.ID.(int64)
		if !ok {
			return false, fmt.Errorf("an invalid server id was provided")
		}

		return a.hasPermissionServer(userID, serverID, requiredFlags)
	}

	return false, fmt.Errorf("an invalid AuthScope.type was provided")
}

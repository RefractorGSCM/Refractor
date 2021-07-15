package authorizer

import "math/big"

func (a *authorizer) hasPermissionServer(userID string, serverID int64, requiredFlags []*big.Int) (bool, error) {
	return false, nil
}

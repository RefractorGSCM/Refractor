package domain

import (
	kratos "github.com/ory/kratos-client-go"
	"math/big"
)

type Traits struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type AuthUser struct {
	Traits *Traits
	*kratos.Session
}

func (au *AuthUser) User() *User {
	return &User{
		Traits: &Traits{
			Email:    au.Traits.Email,
			Username: au.Traits.Username,
		},
		Identity: &au.Identity,
	}
}

type AuthRepo interface {
	CreateUser(userTraits *Traits) (*User, error)
	GetUserByID(id string) (*AuthUser, error)
	GetAllUsers() ([]*AuthUser, error)
}

type AuthObject string

const (
	AuthObjRefractor = AuthObject("REFRACTOR")
	AuthObjServer    = AuthObject("SERVER")
)

type AuthScope struct {
	// Type represents the object type being authenticated against. e.g SERVER
	Type AuthObject

	// ID represents an object identifier. e.g 1 (Server ID)
	ID interface{}
}

// Authorizer represents an entity which can be used to determine if a user has access to perform an action or not.
type Authorizer interface {
	HasPermission(scope AuthScope, userID string, requiredFlags []*big.Int) (bool, error)
}

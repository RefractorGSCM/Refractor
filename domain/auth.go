package domain

import (
	kratos "github.com/ory/kratos-client-go"
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

type Authorizer interface {
	HasPermission(userID, domain, object, action string) (bool, error)
}

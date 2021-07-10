package domain

type Platform interface {
	GetName() string
	GetPlayerIDField() string
}

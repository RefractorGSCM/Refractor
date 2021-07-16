package domain

import (
	"context"
	"time"
)

type Group struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Color       int       `json:"color"`
	Position    int       `json:"position"`
	Permissions string    `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}

type GroupRepo interface {
	Store(ctx context.Context, group *Group) error
	GetAll(ctx context.Context) ([]*Group, error)
	GetByID(ctx context.Context, id int64) (*Group, error)
	GetUserGroups(ctx context.Context, userID string) ([]*Group, error)
}

type GroupService interface {
	Store(c context.Context, group *Group) error
	GetAll(c context.Context) ([]*Group, error)
	GetByID(c context.Context, id int64) (*Group, error)
}

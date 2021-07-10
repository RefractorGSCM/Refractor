package domain

import (
	"context"
	"time"
)

type Server struct {
	ID           int64     `json:"id"`
	Game         string    `json:"game"`
	Name         string    `json:"string"`
	Address      string    `json:"address"`
	RCONPort     uint16    `json:"-"`
	RCONPassword string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type ServerRepo interface {
	Store(ctx context.Context, server *Server) error
	GetByID(ctx context.Context, id int64) (*Server, error)
}

type ServerService interface {
	Store(c context.Context, server *Server) error
	GetByID(c context.Context, id int64) (*Server, error)
}

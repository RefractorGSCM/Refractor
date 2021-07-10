package domain

import "context"

type Server struct {
	ID           int64  `json:"id"`
	Name         string `json:"string"`
	Address      string `json:"address"`
	RCONPort     uint16 `json:"-"`
	RCONPassword string `json:"-"`
}

type ServerRepository interface {
	Store(ctx context.Context, server *Server) error
	GetByID(ctx context.Context, id int64) (*Server, error)
}

type ServerService interface {
	Store(c context.Context, server *Server) error
	GetByID(c context.Context, id int64) error
}

package service

import (
	"Refractor/domain"
	"context"
	"time"
)

type serverService struct {
	repo    domain.ServerRepo
	timeout time.Duration
}

func NewServerService(repo domain.ServerRepo, timeout time.Duration) domain.ServerService {
	return &serverService{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *serverService) Store(c context.Context, server *domain.Server) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Store(ctx, server)
}

func (s *serverService) GetByID(c context.Context, id int64) (*domain.Server, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByID(ctx, id)
}

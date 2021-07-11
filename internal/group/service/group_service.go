package service

import (
	"Refractor/domain"
	"context"
	"time"
)

type groupService struct {
	repo    domain.GroupRepo
	timeout time.Duration
}

func NewGroupService(repo domain.GroupRepo, timeout time.Duration) domain.GroupService {
	return &groupService{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *groupService) Store(c context.Context, group *domain.Group) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Store(ctx, group)
}

func (s *groupService) GetAll(c context.Context) ([]*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetAll(ctx)
}

func (s *groupService) GetByID(c context.Context, id int64) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByID(ctx, id)
}

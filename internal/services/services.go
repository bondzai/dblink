package services

import (
	"context"

	"github.com/bondzai/dblink/internal/interfaces/repositories"
	"github.com/bondzai/dblink/internal/models"
)

type UserService struct {
	repo *repositories.RedisRepository
}

func NewUserService(repo *repositories.RedisRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, location *models.Location) (*models.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Location = *location
	err = s.repo.UpdateUser(ctx, id, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

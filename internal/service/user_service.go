package service

import (
	"auth-go/internal/repository"
	"context"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) RegisterUser(ctx context.Context, username string, password string) error {
	return s.userRepo.RegisterUser(ctx, username, password)
}

func (s *UserService) LoginUser(ctx context.Context, username, password string) (string, error) {
	return s.userRepo.LoginUser(ctx, username, password)
}

func (s *UserService) LogoutUser(ctx context.Context, sessionToken string) (string, error) {
	return s.userRepo.LogoutUser(ctx, sessionToken)
}

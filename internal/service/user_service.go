package service

import "auth-go/internal/repository"

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) RegisterUser(username string, password string) error {
	return s.userRepo.RegisterUser(username, password)
}

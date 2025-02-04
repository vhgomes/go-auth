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

func (s *UserService) LoginUser(username, password string) (string, error) {
	return s.userRepo.LoginUser(username, password)
}

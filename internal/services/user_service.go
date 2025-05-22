package services

import (
	"github.com/thangit93/echo-base/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetAllUsers() ([]repositories.User, error) {
	return s.userRepo.GetAllUsers()
}

func (s *UserService) CreateUser(name string, email string, password string) error {
	return s.userRepo.CreateUser(name, email, password)
}

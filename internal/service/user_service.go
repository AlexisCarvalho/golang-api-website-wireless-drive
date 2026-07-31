package service

import (
	"wireless_drive/internal/model"
	"wireless_drive/internal/repository"
)

type UserService interface {
	RegisterUser(entry *model.User) error
	Authenticate(code, password string) *model.User
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) RegisterUser(entry *model.User) error {
	return s.repo.Create(entry)
}

func (s *userService) Authenticate(code, password string) *model.User {
	user, err := s.repo.FindByCode(code)
	if err != nil || user.Password != password {
		return nil
	}
	return user
}

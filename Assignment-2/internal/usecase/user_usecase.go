package usecase

import (
	"assignment-2/internal/repository"
	"assignment-2/pkg/modules"
	"fmt"
)

// UserUsecase is the business-logic layer that sits between handlers and repository.
type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) GetUsers() ([]modules.User, error) {
	return u.repo.GetUsers()
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id: %d", id)
	}
	return u.repo.GetUserByID(id)
}

func (u *UserUsecase) CreateUser(req modules.CreateUserRequest) (int, error) {
	if req.Name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return 0, fmt.Errorf("email is required")
	}
	return u.repo.CreateUser(req)
}

func (u *UserUsecase) UpdateUser(id int, req modules.UpdateUserRequest) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id: %d", id)
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	return u.repo.UpdateUser(id, req)
}

func (u *UserUsecase) DeleteUser(id int) (int64, error) {
	if id <= 0 {
		return 0, fmt.Errorf("invalid user id: %d", id)
	}
	return u.repo.DeleteUser(id)
}

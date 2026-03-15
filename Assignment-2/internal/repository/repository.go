package repository

import (
	"assignment-2/internal/repository/_postgres"
	"assignment-2/internal/repository/_postgres/users"
	"assignment-2/pkg/modules"
)

// UserRepository defines all database operations for users.
type UserRepository interface {
	GetUsers() ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(req modules.CreateUserRequest) (int, error)
	UpdateUser(id int, req modules.UpdateUserRequest) error
	DeleteUser(id int) (int64, error)
}

type Repositories struct {
	UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}

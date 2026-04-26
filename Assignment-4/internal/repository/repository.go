package repository

import (
	"assignment-4/internal/repository/_postgres"
	"assignment-4/internal/repository/_postgres/users"
	"assignment-4/pkg/modules"
)

type UserRepository interface {
	GetUsers() ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(req modules.CreateUserRequest) (int, error)
	UpdateUser(id int, req modules.UpdateUserRequest) error
	DeleteUser(id int) (int64, error)
	GetPaginatedUsers(f modules.UserFilter) (modules.PaginatedResponse, error)
	GetCommonFriends(userID1, userID2 int) ([]modules.User, error)
}

type Repositories struct {
	UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}

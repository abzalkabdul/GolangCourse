package repo

import (
	"assignment_5/internal/entity"
	"assignment_5/pkg/postgres"
	"fmt"

	"github.com/google/uuid"
)

type UserRepo struct {
	PG *postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (u *UserRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	if err := u.PG.Conn.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepo) LoginUser(user *entity.LoginUserDTO) (*entity.User, error) {
	var userFromDB entity.User
	if err := u.PG.Conn.Where("username = ?", user.Username).First(&userFromDB).Error; err != nil {
		return nil, fmt.Errorf("username not found: %w", err)
	}
	return &userFromDB, nil
}

func (u *UserRepo) GetUserByID(userID uuid.UUID) (*entity.User, error) {
	var user entity.User
	if err := u.PG.Conn.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (u *UserRepo) PromoteUser(userID uuid.UUID) (*entity.User, error) {
	var user entity.User
	if err := u.PG.Conn.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	user.Role = "admin"
	if err := u.PG.Conn.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to promote user: %w", err)
	}
	return &user, nil
}

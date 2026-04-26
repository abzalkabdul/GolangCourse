package usecase

import (
	"assignment_5/internal/entity"

	"github.com/google/uuid"
)

type (
	UserInterface interface {
		RegisterUser(user *entity.User) (*entity.User, string, error)
		LoginUser(user *entity.LoginUserDTO) (string, error)
		GetMe(userID uuid.UUID) (*entity.User, error)
		PromoteUser(userID uuid.UUID) (*entity.User, error)
	}
)

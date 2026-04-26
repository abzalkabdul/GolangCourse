package usecase

import (
	"assignment_5/internal/entity"
	"assignment_5/internal/usecase/repo"
	"assignment_5/utils"
	"fmt"

	"github.com/google/uuid"
)

type UserUseCase struct {
	repo *repo.UserRepo
}

func NewUserUseCase(r *repo.UserRepo) *UserUseCase {
	return &UserUseCase{repo: r}
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
	createdUser, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, "", fmt.Errorf("register user: %w", err)
	}
	sessionID := uuid.New().String()
	return createdUser, sessionID, nil
}

func (u *UserUseCase) LoginUser(user *entity.LoginUserDTO) (string, error) {
	userFromRepo, err := u.repo.LoginUser(user)
	if err != nil {
		return "", fmt.Errorf("user from repo: %w", err)
	}
	if !utils.CheckPassword(userFromRepo.Password, user.Password) {
		return "", fmt.Errorf("invalid credentials")
	}
	token, err := utils.GenerateJWT(userFromRepo.ID, userFromRepo.Role)
	if err != nil {
		return "", fmt.Errorf("generate JWT: %w", err)
	}
	return token, nil
}

func (u *UserUseCase) GetMe(userID uuid.UUID) (*entity.User, error) {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}
	return user, nil
}

func (u *UserUseCase) PromoteUser(userID uuid.UUID) (*entity.User, error) {
	user, err := u.repo.PromoteUser(userID)
	if err != nil {
		return nil, fmt.Errorf("promote user: %w", err)
	}
	return user, nil
}

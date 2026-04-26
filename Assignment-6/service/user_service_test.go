package service

import (
	"assignment-6/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── GetUserByID ──────────────────────────────────────────────────────────────

func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Abzal"}
	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)

	result, err := svc.GetUserByID(1)
	require.NoError(t, err)
	assert.Equal(t, user, result)
}

// ── CreateUser ───────────────────────────────────────────────────────────────

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 2, Name: "Abzal"}
	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := svc.CreateUser(user)
	assert.NoError(t, err)
}

// ── RegisterUser ─────────────────────────────────────────────────────────────

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(m *repository.MockUserRepository, user *repository.User)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "user already exists",
			setup: func(m *repository.MockUserRepository, user *repository.User) {
				existing := &repository.User{ID: 99, Name: "Old"}
				m.EXPECT().GetByEmail("test@example.com").Return(existing, nil)
			},
			wantErr:   true,
			errSubstr: "already exists",
		},
		{
			name: "new user success",
			setup: func(m *repository.MockUserRepository, user *repository.User) {
				m.EXPECT().GetByEmail("test@example.com").Return(nil, nil)
				m.EXPECT().CreateUser(user).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error on CreateUser",
			setup: func(m *repository.MockUserRepository, user *repository.User) {
				m.EXPECT().GetByEmail("test@example.com").Return(nil, nil)
				m.EXPECT().CreateUser(user).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repository.NewMockUserRepository(ctrl)
			svc := NewUserService(mockRepo)
			user := &repository.User{ID: 1, Name: "Abzal", Email: "test@example.com"}

			tt.setup(mockRepo, user)

			err := svc.RegisterUser(user, "test@example.com")
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── UpdateUserName ────────────────────────────────────────────────────────────

func TestUpdateUserName(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		newName   string
		setup     func(m *repository.MockUserRepository)
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "empty name",
			id:        2,
			newName:   "",
			setup:     func(m *repository.MockUserRepository) {},
			wantErr:   true,
			errSubstr: "name cannot be empty",
		},
		{
			name:    "user not found / repo error",
			id:      2,
			newName: "NewName",
			setup: func(m *repository.MockUserRepository) {
				m.EXPECT().GetUserByID(2).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name:    "successful update",
			id:      2,
			newName: "UpdatedName",
			setup: func(m *repository.MockUserRepository) {
				user := &repository.User{ID: 2, Name: "OldName"}
				m.EXPECT().GetUserByID(2).Return(user, nil)
				m.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(u *repository.User) error {
					assert.Equal(t, "UpdatedName", u.Name, "name should be changed before update")
					return nil
				})
			},
			wantErr: false,
		},
		{
			name:    "UpdateUser fails",
			id:      2,
			newName: "UpdatedName",
			setup: func(m *repository.MockUserRepository) {
				user := &repository.User{ID: 2, Name: "OldName"}
				m.EXPECT().GetUserByID(2).Return(user, nil)
				m.EXPECT().UpdateUser(gomock.Any()).Return(errors.New("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repository.NewMockUserRepository(ctrl)
			svc := NewUserService(mockRepo)

			tt.setup(mockRepo)

			err := svc.UpdateUserName(tt.id, tt.newName)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── DeleteUser ────────────────────────────────────────────────────────────────

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		setup     func(m *repository.MockUserRepository)
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "attempt to delete admin",
			id:        1,
			setup:     func(m *repository.MockUserRepository) {},
			wantErr:   true,
			errSubstr: "not allowed to delete admin",
		},
		{
			name: "successful delete",
			id:   2,
			setup: func(m *repository.MockUserRepository) {
				m.EXPECT().DeleteUser(2).DoAndReturn(func(id int) error {
					assert.Equal(t, 2, id, "correct user id should be deleted")
					return nil
				})
			},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   3,
			setup: func(m *repository.MockUserRepository) {
				m.EXPECT().DeleteUser(3).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repository.NewMockUserRepository(ctrl)
			svc := NewUserService(mockRepo)

			tt.setup(mockRepo)

			err := svc.DeleteUser(tt.id)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

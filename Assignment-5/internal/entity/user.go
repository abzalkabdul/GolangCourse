package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Username string    `json:"username" gorm:"uniqueIndex;not null"`
	Email    string    `json:"email" gorm:"uniqueIndex;not null"`
	Password string    `json:"-" gorm:"not null"`
	Role     string    `json:"role" gorm:"default:user"`
	Verified bool      `json:"verified" gorm:"default:false"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

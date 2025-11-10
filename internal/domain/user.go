package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null"`
	Password  string     `json:"-" gorm:"not null"`
	Name      string     `json:"name"`
	Surname   string     `json:"surname"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	Roles     []UserRole `json:"roles" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type UserRole struct {
	ID     uint      `json:"-" gorm:"primaryKey"`
	UserID uuid.UUID `json:"-" gorm:"type:uuid;not null"`
	Role   string    `json:"role" gorm:"not null"`
	User   User      `json:"-" gorm:"foreignKey:UserID"`
}

const (
	RoleCustomer = "customer"
	RoleSeller   = "seller"
)

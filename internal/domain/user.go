package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	Phone     string     `json:"phone" gorm:"uniqueIndex;not null"`
	Password  string     `json:"-" gorm:"not null"`
	Name      string     `json:"name"`
	Surname   string     `json:"surname"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	Roles     []UserRole `json:"roles" gorm:"foreignKey:UserID"`
}

// BeforeCreate generates a UUID for the user before persisting
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID     uint      `json:"-" gorm:"primaryKey"`
	UserID uuid.UUID `json:"-" gorm:"type:uuid;not null"`
	Role   string    `json:"role" gorm:"not null"`
	User   User      `json:"-" gorm:"foreignKey:UserID"`
}

// Role constants
const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
	RoleSeller   = "seller"
)

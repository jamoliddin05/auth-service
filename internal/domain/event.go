package domain

import (
	"time"
)

type Event struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;type:bigint"`
	Type      string    `gorm:"not null"`
	Payload   string    `gorm:"type:jsonb;not null"`
	Processed bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
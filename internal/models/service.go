package models

import (
	"time"
)

type Service struct {
	ID         uint   `gorm:"primaryKey"`
	RoomID     string `gorm:"not null;unique"`
	IsResolved bool   `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

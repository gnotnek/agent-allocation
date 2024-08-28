package models

import "time"

type Service struct {
	ID              uint   `gorm:"primaryKey"`
	RoomID          string `gorm:"not null"`
	IsResolved      bool   `gorm:"not null"`
	Notes           string
	FirstCommentID  string
	LastCommentID   string
	Source          string
	ResolvedBy      string
	ResolvedByEmail string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

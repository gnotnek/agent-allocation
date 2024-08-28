package models

import "time"

type Customer struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"not null"`
	RoomID    string    `gorm:"not null"`
	Status    string    `gorm:"not null"` // waiting, served
	CreatedAt time.Time // FIFO based on this timestamp
}

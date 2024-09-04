package models

import "time"

type RoomQueue struct {
	ID        uint   `json:"id"`
	RoomID    string `json:"room_id" gorm:"unique"`
	Position  int    `json:"position"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

package models

import "time"

type RoomQueue struct {
	ID        uint      `json:"id"`
	RoomID    string    `json:"room_id" gorm:"unique"`
	AgentID   uint      `json:"agent_id"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

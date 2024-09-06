package models

import (
	"time"
)

type RoomQueue struct {
	ID        uint   `json:"id"`
	RoomID    string `json:"room_id" gorm:"uniqueIndex"`
	AgentID   *int   `json:"agent_id,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

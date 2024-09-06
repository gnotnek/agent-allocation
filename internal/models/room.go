package models

import (
	"time"
)

type RoomQueue struct {
	ID        uint   `json:"id"`
	RoomID    string `json:"room_id"`
	AgentID   *int   `json:"agent_id,omitempty"` // Null if unassigned
	CreatedAt time.Time
	UpdatedAt time.Time
}

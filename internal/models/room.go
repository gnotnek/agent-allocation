package models

import (
	"time"
)

type RoomQueue struct {
	ID         uint   `json:"id"`
	RoomID     string `json:"room_id"`
	AgentID    *int   `json:"agent_id,omitempty"` // Null if unassigned
	IsResolved bool   `json:"is_resolved"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

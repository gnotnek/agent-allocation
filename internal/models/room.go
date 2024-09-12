package models

import (
	"time"
)

type RoomQueue struct {
	ID         uint      `json:"id"`
	RoomID     string    `json:"room_id" gorm:"uniqueIndex"`
	AgentID    *int      `json:"agent_id,omitempty"`
	CreatedAt  time.Time `gorm:"index"`
	AssignedAt time.Time `json:"assigned_at,omitempty"`
	ResolvedAt time.Time `json:"resolved_at,omitempty"`
	UpdatedAt  time.Time
}

// bikin colom baru untuk waktu agent id sama menyelesaikan room

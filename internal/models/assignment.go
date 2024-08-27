package models

import (
	"gorm.io/gorm"
)

type Assignment struct {
	gorm.Model
	AgentID    uint     `json:"agent_id"`
	CustomerID string   `json:"customer_id"`
	Agent      Agent    `json:"agent" gorm:"foreignKey:AgentID"`
	Customer   Customer `json:"customer" gorm:"foreignKey:CustomerID"`
}

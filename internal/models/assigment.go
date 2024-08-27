package models

type Assigment struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	AgentID    uint     `json:"agent_id"`
	CustomerID uint     `json:"customer_id"`
	Agent      Agent    `json:"agent"`
	Customer   Customer `json:"customer"`
}

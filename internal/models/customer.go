package models

type Customer struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	GroupID   uint   `json:"group_id"`
	Status    string `json:"status"`
	Priyority int    `json:"priyority"`
}

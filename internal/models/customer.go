package models

type Customer struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	RoomID string `json:"room_id"`
	AppID  string `json:"agent_id"`
	Status string `json:"status"`
	Extras string `gorm:"type:json" json:"extras"`
}

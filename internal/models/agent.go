package models

type Agent struct {
	ID               uint     `gorm:"primaryKey"`
	Name             string   `gorm:"not null"`
	Email            string   `gorm:"unique;not null"`
	Status           string   `gorm:"not null"` // online, offline
	MaxCustomers     int      `gorm:"not null"`
	CurrentCustomers int      `gorm:"not null"`
	AssignedRules    []string `gorm:"type:text[]"`
}

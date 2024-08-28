package models

import (
	"gorm.io/gorm"
)

type Agent struct {
	gorm.Model
	ID              uint   `json:"id" gorm:"primaryKey"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	MaxCustomer     int    `json:"max_customer"`
	CurrentCustomer int    `json:"current_customer"`
	Specialization  string `json:"specialization"`
}

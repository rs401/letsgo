package models

import "gorm.io/gorm"

// Image model for user uploaded image
type Image struct {
	gorm.Model
	FileName string
	UserID   uint `json:"userid" gorm:"not null"`
	User     User
}

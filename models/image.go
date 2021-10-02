package models

import "gorm.io/gorm"

type Image struct {
	gorm.Model
	FileName string
	UserID   uint `json:"userid" gorm:"not null"`
	User     User
}

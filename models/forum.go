package models

import "gorm.io/gorm"

type Forum struct {
	gorm.Model
	Name    string `json:"name" gorm:"not null" form:"name"`
	UserID  uint   `json:"userid" gorm:"not null"`
	User    User
	Threads []Thread
}

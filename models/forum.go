package models

import "gorm.io/gorm"

type Forum struct {
	gorm.Model
	Name    string `json:"name" gorm:"not null" form:"name"`
	UserID  uint   `json:"userid" gorm:"not null"`
	User    User
	Threads []Thread
}

type NewForum struct {
	Name string `json:"name" form:"name"`
	Csrf string `json:"csrf" form:"csrf"`
}

type DelForum struct {
	Csrf string `json:"csrf" form:"csrf"`
}

package models

import "gorm.io/gorm"

type Forum struct {
	gorm.Model
	Name        string `json:"name" gorm:"not null" form:"name"`
	Description string `json:"description" gorm:"size:256"`
	Open        bool   `json:"Open" gorm:"default:true"`
	UserID      uint   `json:"userid" gorm:"not null"`
	User        User
	Threads     []Thread
	Members     []Member
}

type NewForum struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Csrf        string `json:"csrf" form:"csrf"`
}

type UpdateForum struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Open        bool   `json:"Open" form:"open"`
	Csrf        string `json:"csrf" form:"csrf"`
}

type DelForum struct {
	Csrf string `json:"csrf" form:"csrf"`
}

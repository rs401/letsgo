package models

import "gorm.io/gorm"

// Forum model
type Forum struct {
	gorm.Model
	Name        string `json:"name" gorm:"not null" form:"name"`
	Description string `json:"description" gorm:"size:256"`
	Open        bool   `json:"Open" gorm:"default:true"`
	UserID      uint   `json:"userid" gorm:"not null"`
	User        User
	Threads     []Thread
	Members     []Member
	Tags        []Tag `gorm:"many2many:forum_tags;"`
}

// NewForum form model
type NewForum struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Csrf        string `json:"csrf" form:"csrf"`
}

// UpdateForum form model
type UpdateForum struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Open        bool   `json:"Open" form:"open"`
	Csrf        string `json:"csrf" form:"csrf"`
}

// DelForum form model
type DelForum struct {
	Csrf string `json:"csrf" form:"csrf"`
}

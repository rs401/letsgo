package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	DisplayName string `json:"display_name" gorm:"unique;not null"`
	Email       string `json:"email" gorm:"unique;not null"`
	Password    []byte `json:"-"`
	Posts       []Post
	Threads     []Thread
	Forums      []Forum
}

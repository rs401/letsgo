package models

import "gorm.io/gorm"

type Thread struct {
	gorm.Model
	Title   string `json:"title" gorm:"not null"`
	Body    string `json:"body" gorm:"not null"`
	UserID  uint   `json:"userid" gorm:"not null"`
	User    User
	ForumID uint `json:"forum_id" gorm:"not null"`
	Forum   Forum
	Posts   []Post `json:""`
}

type NewThread struct {
	Title string `json:"title" form:"title"`
	Body  string `json:"body" form:"body"`
	Csrf  string `json:"csrf" form:"csrf"`
}

package models

import (
	"time"

	"gorm.io/gorm"
)

// Thread model for thread
type Thread struct {
	gorm.Model
	Title   string    `json:"title" gorm:"not null"`
	Body    string    `json:"body" gorm:"not null"`
	Date    time.Time `json:"date" `
	UserID  uint      `json:"userid" gorm:"not null"`
	User    User
	ForumID uint `json:"forum_id" gorm:"constraint:OnDelete:CASCADE;"`
	Forum   Forum
	Posts   []Post `json:""`
}

// NewThread form model for new thread
type NewThread struct {
	Title string `json:"title" form:"title"`
	Body  string `json:"body" form:"body"`
	Csrf  string `json:"csrf" form:"csrf"`
}

// UpdateThread form model for update thread
type UpdateThread struct {
	Title string    `json:"title" form:"title"`
	Body  string    `json:"body" form:"body"`
	Csrf  string    `json:"csrf" form:"csrf"`
	Date  time.Time `json:"date" form:"date"`
}

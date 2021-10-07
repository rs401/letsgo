package models

import "gorm.io/gorm"

// Member model for forum member
type Member struct {
	gorm.Model
	ForumID uint `json:"ForumID" gorm:"not null"`
	Forum   Forum
	UserID  uint `json:"UserID" gorm:"not null"`
	User    User
}

// PendingMember model for pending member requests to join private forum
type PendingMember struct {
	gorm.Model
	ForumID uint `json:"ForumID" gorm:"not null"`
	Forum   Forum
	UserID  uint `json:"UserID" gorm:"not null"`
	User    User
}

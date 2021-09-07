package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	Body     string `json:"body" gorm:"not null"`
	UserID   uint   `json:"userid" gorm:"not null"`
	User     User
	ThreadID uint `json:"thread_id" gorm:"not null"`
	Thread   Thread
	// Thread Thread //`json:"thread"`
}

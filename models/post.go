package models

import "gorm.io/gorm"

// Post model for post reply in a thread
type Post struct {
	gorm.Model
	Body     string `json:"body" gorm:"not null"`
	UserID   uint   `json:"userid" gorm:"not null"`
	User     User
	ThreadID uint `json:"thread_id" gorm:"not null"`
	Thread   Thread
	// Thread Thread //`json:"thread"`
}

// NewPost model for new post form
type NewPost struct {
	Body string `json:"body" form:"body"`
	Csrf string `json:"csrf" form:"csrf"`
}

package models

import "gorm.io/gorm"

// User model for user
type User struct {
	gorm.Model
	DisplayName string   `json:"DisplayName" gorm:"unique;not null"`
	Email       string   `json:"email" gorm:"unique;not null"`
	Picture     string   `json:"picture"`
	Password    []byte   `json:"-"`
	Posts       []Post   `json:"-"`
	Threads     []Thread `json:"-"`
	Forums      []Forum  `json:"-"`
}

// GUser model for Google Sign-in user information
type GUser struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Email   string `json:"email"`
}

// NewUser form model for new user registration
type NewUser struct {
	DisplayName string `json:"DisplayName"`
	Email       string `json:"email"`
	Pass1       string `json:"pass1"`
	Pass2       string `json:"pass2"`
	Csrf        string `json:"csrf" form:"csrf"`
}

// LoginUser form model for user login
type LoginUser struct {
	Email string `json:"email"`
	Pass1 string `json:"pass1"`
	Csrf  string `json:"csrf" form:"csrf"`
}

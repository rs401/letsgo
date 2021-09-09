package models

import "gorm.io/gorm"

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

type GUser struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Email   string `json:"email"`
}

type NewUser struct {
	DisplayName string `json:"DisplayName"`
	Email       string `json:"email"`
	Pass1       string `json:"pass1"`
	Pass2       string `json:"pass2"`
}

type LoginUser struct {
	Email string `json:"email"`
	Pass1 string `json:"pass1"`
}

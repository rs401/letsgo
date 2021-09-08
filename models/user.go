package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	DisplayName string   `json:"display_name" gorm:"unique;not null"`
	Email       string   `json:"email" gorm:"unique;not null"`
	Password    []byte   `json:"-"`
	Posts       []Post   `json:"-"`
	Threads     []Thread `json:"-"`
	Forums      []Forum  `json:"-"`
}

type NewUser struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Pass1       string `json:"pass1"`
	Pass2       string `json:"pass2"`
}

type LoginUser struct {
	Email string `json:"email"`
	Pass1 string `json:"pass1"`
}

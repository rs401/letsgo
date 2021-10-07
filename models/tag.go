package models

import "gorm.io/gorm"

// Tag is a keyword tag for groups
type Tag struct {
	gorm.Model
	Name string
}

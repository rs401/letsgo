package models

// import (
// 	"time"

// 	"gorm.io/gorm"
// )

// type Recipe struct {
// 	gorm.Model
// 	Name         string        `json:"Name"`
// 	Tags         []Tag         `json:"Tags"`
// 	Ingredients  []Ingredient  `json:"Ingredients"`
// 	Instructions []Instruction `json:"Instructions"`
// 	PublishedAt  time.Time     `json:"PublishedAt"`
// }

// type Tag struct {
// 	gorm.Model
// 	Name     string `json:"Name"`
// 	RecipeID uint   `json:"RecipeID" gorm:"not null"`
// 	Recipe   Recipe
// }

// type Ingredient struct {
// 	gorm.Model
// 	Name     string `json:"Name"`
// 	RecipeID uint   `json:"RecipeID" gorm:"not null"`
// 	Recipe   Recipe
// }

// type Instruction struct {
// 	gorm.Model
// 	Name     string `json:"Name"`
// 	RecipeID uint   `json:"RecipeID" gorm:"not null"`
// 	Recipe   Recipe
// }

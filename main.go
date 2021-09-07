// Recipes API
//
// This is a sample recipes API. You can find out more about the API at https://github.com/rs401/letsgo/.
//
// Schemes: http
// Host: 127.0.0.1:9000
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs401/letsgo/handlers"
	"github.com/rs401/letsgo/models"
)

// var recipes []models.Recipe

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	models.InitDatabase()
	// db := models.DBConn
	// recipes = make([]models.Recipe, 0)
	// file, _ := ioutil.ReadFile("./recipes.json")
	// _ = json.Unmarshal([]byte(file), &recipes)
	// for _, recipe := range recipes {
	// 	var rec models.Recipe
	// 	rec = recipe
	// 	db.Create(&rec)
	// }
	// db.Create(&recipes)
}

func Config(key string) string {
	return os.Getenv(key)
}

func main() {
	r := gin.Default()
	r.GET("/", handlers.IndexHandler)
	// r.POST("/recipes", handlers.NewRecipeHandler)
	// r.GET("/recipes", handlers.ListRecipesHandler)
	// r.PUT("/recipes/:id", handlers.UpdateRecipeHandler)
	// r.DELETE("/recipes/:id", handlers.DeleteRecipeHandler)
	// r.GET("/recipes/search", handlers.SearchRecipesHandler)
	r.GET("/forums", handlers.GetForums)
	r.POST("/forums", handlers.NewForum)
	r.GET("/forums/:id", handlers.GetForum)
	r.PUT("/forums/:id", handlers.UpdateForum)
	r.DELETE("/forums/:id", handlers.DeleteForum)

	r.Run(":" + Config("API_PORT"))
}

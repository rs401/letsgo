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

	// Seed some forums for testing
	// bytesRead, _ := ioutil.ReadFile("words.txt")
	// file_content := string(bytesRead)
	// words := strings.Split(file_content, "\n")
	// db := models.DBConn
	// for _, name := range words {
	// 	forum := models.Forum{Name: name}
	// 	db.Create(&forum)
	// }

}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X_API_KEY") != os.Getenv("X_API_KEY") {
			c.AbortWithStatus(401)
		}
		c.Next()
	}
}

func Config(key string) string {
	return os.Getenv(key)
}

func main() {
	r := gin.Default()
	r.GET("/", handlers.IndexHandler)
	r.GET("/forums", handlers.GetForums)
	authorized := r.Group("/")
	authorized.Use(AuthMiddleware())
	{
		authorized.POST("/forums", handlers.NewForum)
		authorized.GET("/forums/:id", handlers.GetForum)
		authorized.PUT("/forums/:id", handlers.UpdateForum)
		authorized.DELETE("/forums/:id", handlers.DeleteForum)

	}
	// r.POST("/recipes", handlers.NewRecipeHandler)
	// r.GET("/recipes", handlers.ListRecipesHandler)
	// r.PUT("/recipes/:id", handlers.UpdateRecipeHandler)
	// r.DELETE("/recipes/:id", handlers.DeleteRecipeHandler)
	// r.GET("/recipes/search", handlers.SearchRecipesHandler)

	r.Run(":" + Config("API_PORT"))
}

package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
}

func Config(key string) string {
	return os.Getenv(key)
}

func IndexHandler(c *gin.Context) {
	name := c.Params.ByName("name")
	c.JSON(200, gin.H{
		"message": "hello," + name,
	})
}

func main() {
	port := Config("API_PORT")
	router := gin.Default()
	router.GET("/:name", IndexHandler)

	router.Run(":" + port)
}

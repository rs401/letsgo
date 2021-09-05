package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Recipe struct {
	Name         string    `json:"Name"`
	Tags         []string  `json:"Tags"`
	Ingredients  []string  `json:"Ingredients"`
	Instructions []string  `json:"Instructions"`
	PublishedAt  time.Time `json:"PublishedAt"`
}

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
}

func Config(key string) string {
	return os.Getenv(key)
}

func IndexHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Let's Go",
	})
}

func main() {
	port := Config("API_PORT")
	router := gin.Default()
	router.GET("/", IndexHandler)

	router.Run(":" + port)
}

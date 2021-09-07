package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/rs401/letsgo/models"
)

// Forums
// Get all forums
func GetForums(c *gin.Context) {
	db := models.DBConn
	redisClient := models.RedisClient
	var forums []models.Forum
	val, err := redisClient.Get(c, "forums").Result()
	if err == redis.Nil {
		log.Printf("==== Not cached, Querying db")
		result := db.Find(&forums)
		if result.Error != nil {
			log.Printf("==== Error Querying db")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Error",
			})
			return
		}
		data, _ := json.Marshal(forums)
		redisClient.Set(c, "forums", string(data), 0)
		c.JSON(http.StatusOK, forums)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Error",
		})
		return
	} else {
		log.Printf("==== Request to Redis")
		forums = make([]models.Forum, 0)
		json.Unmarshal([]byte(val), &forums)
		c.JSON(http.StatusOK, forums)
	}
}

// Get single forum
func GetForum(c *gin.Context) {
	db := models.DBConn
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Badrequest",
		})
		return
	}
	var forum models.Forum
	db.Preload("Threads").Find(&forum, id)
	if forum.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "notfound",
		})
		return
	}
	c.JSON(http.StatusOK, forum)
}

func NewForum(c *gin.Context) {
	if c.GetHeader("X_API_KEY") != os.Getenv("X_API_KEY") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "API Key not provided or invalid.",
		})
		return
	}
	db := models.DBConn
	forum := new(models.Forum)
	if err := c.ShouldBindJSON(&forum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "badrequest",
		})
		return
	}
	result := db.Create(&forum)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, "forums")
	c.JSON(http.StatusOK, forum)
}

func DeleteForum(c *gin.Context) {
	// Grab db
	db := models.DBConn
	// Convert string parameter to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "badrequest",
		})
		return
	}
	// Get the forum
	var forum models.Forum
	res := db.Find(&forum, id)
	if res.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "notfound",
		})
		return
	}
	if forum.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "notfound",
		})
		return
	}
	db.Delete(&forum)
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, "forums")
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func UpdateForum(c *gin.Context) {
	// Grab db
	db := models.DBConn
	// Convert string parameter to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "badrequest",
		})
		return
	}
	// Get original and apply updates
	var forum models.Forum
	var updForum = new(models.Forum)
	res := db.First(&forum, id)
	if res.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "notfound",
		})
		return
	}
	// Parse new values
	if err := c.ShouldBindJSON(updForum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "badrequest",
		})
		return
	}
	// Check not empty string
	if updForum.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "badrequest",
		})
		return
	}
	// Update forum and save
	forum.Name = updForum.Name
	db.Save(&forum)
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, "forums")

	c.JSON(http.StatusOK, forum)
}

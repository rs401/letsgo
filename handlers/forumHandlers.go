package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs401/letsgo/models"
)

// Forums
// Get all forums
func GetForums(c *gin.Context) {
	db := models.DBConn
	var forums []models.Forum
	db.Find(&forums)
	c.JSON(http.StatusOK, forums)
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
	db := models.DBConn
	forum := new(models.Forum)
	if err := c.ShouldBindJSON(&forum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "badrequest",
		})
		return
	}
	db.Create(&forum)
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

	c.JSON(http.StatusOK, forum)
}

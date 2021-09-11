package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/rs401/letsgo/models"
)

type ForumHandler struct{}

// Forums
// Get all forums
func (handler *ForumHandler) GetForumsHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))

	db := models.DBConn
	redisClient := models.RedisClient
	var forums []models.Forum
	val, err := redisClient.Get(c, "forums").Result()
	if err == redis.Nil {
		log.Printf("==== Not cached, Querying db")
		result := db.Find(&forums)
		if result.Error != nil {
			session.AddFlash("Internal Error")
			log.Printf("==== Error Querying db")
			c.HTML(http.StatusInternalServerError, "forums.html", gin.H{
				"error":   "Internal Error",
				"flashes": session.Flashes(),
				"user":    email,
			})
			session.Save()
			return
		}
		data, _ := json.Marshal(forums)
		redisClient.Set(c, "forums", string(data), 0)
		c.HTML(http.StatusOK, "forums.html", gin.H{
			"forums": forums,
			"user":   email,
		})
		return
	} else if err != nil {
		session.AddFlash("Internal Error")
		c.HTML(http.StatusInternalServerError, "forums.html", gin.H{
			"error":   "Internal Error",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	} else {
		log.Printf("==== Request to Redis")
		forums = make([]models.Forum, 0)
		json.Unmarshal([]byte(val), &forums)
		c.HTML(http.StatusOK, "forums.html", gin.H{
			"forums": forums,
			"user":   email,
		})
		return
	}
}

// Get single forum
func (handler *ForumHandler) GetForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))

	db := models.DBConn
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "forum.html", gin.H{
			"error": "Badrequest",
			"user":  email,
		})
		return
	}
	var forum models.Forum
	db.Preload("Threads").Preload("User").Find(&forum, id)
	if forum.ID == 0 {
		c.HTML(http.StatusNotFound, "forum.html", gin.H{
			"message": "notfound",
			"user":    email,
		})
		return
	}
	c.HTML(http.StatusOK, "forum.html", gin.H{
		"forum":      forum,
		"forum.User": forum.User,
		"user":       email,
	})
}

func (handler *ForumHandler) NewForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	user := getUserByEmail(email)
	if user == nil {
		c.Redirect(http.StatusUnauthorized, "/login")
		return
	}

	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "new_forum.html", gin.H{
			"user": email,
		})
		return
	}

	db := models.DBConn
	newForum := new(models.NewForum)
	forum := new(models.Forum)
	if err := c.Bind(&newForum); err != nil {
		session.AddFlash("Bad Request")
		c.Redirect(http.StatusFound, "/new_forum")
		session.Save()
		return
	}
	forum.Name = newForum.Name
	forum.User = *user
	result := db.Create(&forum)
	if result.Error != nil {
		session.AddFlash("Internal Error")
		c.Redirect(http.StatusInternalServerError, "/new_forum")
		session.Save()
		return
	}
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, "forums")
	c.Redirect(http.StatusFound, "/forums/"+strconv.Itoa(int(forum.ID)))
}

func (handler *ForumHandler) ConfirmDeleteForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	c.HTML(http.StatusOK, "confirm_delete.html", gin.H{
		"user": email,
		"id":   c.Param("id"),
	})
}

func (handler *ForumHandler) DeleteForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	user := getUserByEmail(email)
	// Grab db
	db := models.DBConn
	// Convert string parameter to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		session.AddFlash("Bad Request")
		c.HTML(http.StatusBadRequest, "confirm_delete.html", gin.H{
			"message": "badrequest",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	}
	// Get the forum
	var forum models.Forum
	res := db.Preload("User").Find(&forum, id)
	if res.Error != nil {
		session.AddFlash("Forum not found")
		c.HTML(http.StatusNotFound, "confirm_delete.html", gin.H{
			"message": "notfound",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	}
	if forum.ID == 0 {
		session.AddFlash("Forum not found")
		c.HTML(http.StatusNotFound, "confirm_delete.html", gin.H{
			"message": "notfound",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	}
	if forum.User.Email != user.Email {
		session.AddFlash("Forbidden")
		c.HTML(http.StatusForbidden, "confirm_delete.html", gin.H{
			"message": "forbidden",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	}
	db.Delete(&forum)
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, "forums")
	c.Redirect(http.StatusFound, "/")
}

func (handler *ForumHandler) UpdateForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	user := getUserByEmail(email)
	if user == nil {
		c.Redirect(http.StatusUnauthorized, "/login")
		return
	}

	// Grab db
	db := models.DBConn
	// Convert string parameter to int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		session.AddFlash("Bad Request")
		c.HTML(http.StatusBadRequest, "update_forum.html", gin.H{
			"message": "badrequest",
			"user":    email,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	// Get original and apply updates
	var forum models.Forum
	var updForum = new(models.NewForum)
	res := db.Preload("User").Find(&forum, id)
	if res.Error != nil {
		session.AddFlash("Group not found")
		c.HTML(http.StatusNotFound, "update_forum.html", gin.H{
			"message": "notfound",
			"user":    email,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "update_forum.html", gin.H{
			"user":  email,
			"id":    id,
			"forum": forum,
		})
		return
	}
	// Parse new values
	if err := c.Bind(updForum); err != nil {
		session.AddFlash("Bad Request")
		c.HTML(http.StatusBadRequest, "update_forum.html", gin.H{
			"message": "badrequest",
			"user":    email,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	// Check not empty string
	if updForum.Name == "" {
		session.AddFlash("Bad Request")
		c.HTML(http.StatusBadRequest, "update_forum.html", gin.H{
			"message": "badrequest",
			"user":    email,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	if forum.User.Email != user.Email {
		session.AddFlash("Forbidden")
		c.HTML(http.StatusForbidden, "update_forum.html", gin.H{
			"message": "forbidden",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	}
	// Update forum and save
	forum.Name = updForum.Name
	db.Save(&forum)
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, "forums")

	c.Redirect(http.StatusFound, "/forums/"+c.Param("id"))
}

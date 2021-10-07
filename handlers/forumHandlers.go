package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs401/letsgo/models"
)

// ForumHandler handler func receiver
type ForumHandler struct{}

// GetForumsHandler returns all forums
func (handler *ForumHandler) GetForumsHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	db := models.DBConn
	redisClient := models.RedisClient
	var forums []models.Forum
	val, err := redisClient.Get(c, "forums").Result()
	if err == redis.Nil {
		log.Printf("==== Not cached, Querying db")
		result := db.Order("created_at desc").Find(&forums)
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

// GetForumHandler returns a single forum
func (handler *ForumHandler) GetForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "forum.html", gin.H{
			"error": "Badrequest",
			"user":  email,
		})
		return
	}

	db := models.DBConn
	var forum models.Forum
	user := getUserByEmail(email)
	db.Preload("User").Find(&forum, id)

	// Check if forum is Open
	if !forum.Open {
		// If not Open, see if User is a member
		if !IsMember(uint(id), user.ID) {
			// If not member show a "Join group" form where the threads would be.
			session.AddFlash("Not a member")
			log.Printf("==== Not a member of group")
			c.HTML(http.StatusOK, "forum.html", gin.H{
				"error":   "Internal Error",
				"flashes": session.Flashes(),
				"user":    email,
				"forum":   forum,
				"id":      id,
				"uid":     user.ID,
				"member":  false,
			})
			session.Save()
			return
		}
	}
	// Set a flag in the response indicating member true/false.
	// Else continue checking redis and whatnot

	// Check redis
	redisClient := models.RedisClient
	val, err := redisClient.Get(c, fmt.Sprintf("forum%v", c.Param("id"))).Result()
	// If redis nil, get from db and set in redis
	if err == redis.Nil {
		log.Printf("==== Not cached, Querying db")
		result := db.Preload("Threads").Preload("User").Order("created_at desc").Find(&forum, id)
		if result.Error != nil {
			session.AddFlash("Internal Error")
			log.Printf("==== Error Querying db")
			c.HTML(http.StatusInternalServerError, "forum.html", gin.H{
				"error":   "Internal Error",
				"flashes": session.Flashes(),
				"user":    email,
			})
			session.Save()
			return
		}
		if forum.ID == 0 {
			session.AddFlash("Not found.")
			c.HTML(http.StatusNotFound, "forum.html", gin.H{
				"message": "notfound",
				"user":    email,
				"flashes": session.Flashes(),
			})
			session.Save()
			return
		}
		data, _ := json.Marshal(forum)
		redisClient.Set(c, fmt.Sprintf("forum%v", c.Param("id")), string(data), 0)
		c.HTML(http.StatusOK, "forum.html", gin.H{
			"forum":      forum,
			"forum.User": forum.User,
			"threads":    forum.Threads,
			"user":       email,
			"member":     true,
		})
		return
	} else if err != nil {
		// Else if != nil, internal error
		session.AddFlash("Internal Error")
		c.HTML(http.StatusInternalServerError, "forum.html", gin.H{
			"error":   "Internal Error",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	} else {
		// Else get it from redis
		log.Printf("==== Request to Redis")
		json.Unmarshal([]byte(val), &forum)
		c.HTML(http.StatusOK, "forum.html", gin.H{
			"forum":      forum,
			"forum.User": forum.User,
			"threads":    forum.Threads,
			"user":       email,
			"member":     true,
		})
		return
	}
}

// IsMember checks if user is a member of the private forum
func IsMember(fid, uid uint) bool {
	var members []models.Member
	db := models.DBConn
	db.Where("forum_id = ?", fid).Find(&members)
	for _, member := range members {
		if member.UserID == uid {
			return true
		}
	}
	return false
}

// NewForumHandler handles returning new forum template on GET and creates new
// forum on POST
func (handler *ForumHandler) NewForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	user := getUserByEmail(email)
	if user == nil {
		c.Redirect(http.StatusUnauthorized, "/login")
		return
	}

	if c.Request.Method == "GET" {
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusOK, "new_forum.html", gin.H{
			"user": email,
			"csrf": csrf,
		})
		session.Save()
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
	// Check csrf
	if newForum.Csrf != fmt.Sprintf("%v", session.Get("csrf")) {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusOK, "new_forum.html", gin.H{
			"user":    email,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	if strings.TrimSpace(newForum.Name) == "" || strings.TrimSpace(newForum.Description) == "" {
		session.AddFlash("Name and Description must not be empty.")
		c.HTML(http.StatusOK, "new_forum.html", gin.H{
			"user":    email,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	forum.Name = newForum.Name
	forum.Description = newForum.Description
	forum.User = *user
	forum.Open = true
	log.Printf("Created New Forum: %v", forum.Name)
	result := db.Create(&forum)
	if result.Error != nil {
		session.AddFlash("Internal Error")
		c.Redirect(http.StatusInternalServerError, "/new_forum")
		session.Save()
		return
	}
	log.Printf("Adding user to member list")
	var member *models.Member = new(models.Member)
	member.ForumID = forum.ID
	member.UserID = user.ID
	db.Create(&member)
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, "forums")
	c.Redirect(http.StatusFound, "/forums/"+strconv.Itoa(int(forum.ID)))
}

// ConfirmDeleteForumHandler
func (handler *ForumHandler) ConfirmDeleteForumHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	csrf := uuid.NewString()
	session.Set("csrf", csrf)
	c.HTML(http.StatusOK, "confirm_delete.html", gin.H{
		"user": email,
		"id":   c.Param("id"),
		"csrf": csrf,
	})
	session.Save()
}

// DeleteForumHandler
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
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}
	var delForum models.DelForum
	if err := c.Bind(&delForum); err != nil {
		session.AddFlash("Bad Request")
		c.Redirect(http.StatusFound, "/forums")
		session.Save()
		return
	}
	// Check csrf
	if delForum.Csrf != fmt.Sprintf("%v", session.Get("csrf")) {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusOK, "confirm_delete.html", gin.H{
			"user":    email,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
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

// UpdateForumHandler
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
	var updForum = new(models.UpdateForum)
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
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusOK, "update_forum.html", gin.H{
			"user":  email,
			"id":    id,
			"forum": forum,
			"csrf":  csrf,
		})
		session.Save()
		return
	}
	// Parse new values
	updForum.Name = c.Request.FormValue("name")
	updForum.Description = c.Request.FormValue("description")
	updForum.Csrf = c.Request.FormValue("csrf")
	open := c.Request.FormValue("open")
	if open == "" {
		updForum.Open = false
	} else {
		updForum.Open = true
	}

	// Check not empty string
	if updForum.Name == "" || updForum.Description == "" {
		session.AddFlash("Bad Request")
		session.AddFlash("Name and Description must not be empty.")
		c.HTML(http.StatusBadRequest, "update_forum.html", gin.H{
			"message": "badrequest",
			"user":    email,
			"id":      id,
			"forum":   forum,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	if forum.User.Email != user.Email {
		session.AddFlash("Forbidden")
		log.Printf("Unauthorized update attempt: %s", user.Email)
		c.HTML(http.StatusForbidden, "update_forum.html", gin.H{
			"message": "forbidden",
			"flashes": session.Flashes(),
			"user":    email,
		})
		session.Save()
		return
	}
	// Check csrf
	if updForum.Csrf != fmt.Sprintf("%v", session.Get("csrf")) {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusOK, "update_forum.html", gin.H{
			"user":    email,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	// Update forum and save
	forum.Name = updForum.Name
	forum.Description = updForum.Description
	forum.Open = updForum.Open
	db.Save(&forum)
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, fmt.Sprintf("forum%v", c.Param("id")))

	c.Redirect(http.StatusFound, "/forums/"+c.Param("id"))
}

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs401/letsgo/models"
)

type ThreadHandler struct{}

// Not needed because I don't want to list all threads without their parent forum
// Get all threads
// func (handler *ThreadHandler) GetThreads(c *gin.Context) {
// 	db := models.DBConn
// 	var threads []models.Thread
// 	fid, err := strconv.Atoi(c.Param("fid"))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}
// 	db.Where(&models.Thread{ForumID: uint(fid)}).Find(&threads)
// 	c.JSON(http.StatusOK, gin.H{
// 		"threads": threads,
// 	})
// 	return
// }

// Get single thread
func (handler *ThreadHandler) GetThreadHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))

	db := models.DBConn
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "thread.html", gin.H{
			"error": "Badrequest",
			"user":  email,
		})
		return
	}

	var thread models.Thread
	db.Preload("Posts").Preload("User").Find(&thread, id)
	if thread.ID == 0 {
		session.AddFlash("Thread does not exist in the database.")
		c.HTML(http.StatusNotFound, "thread.html", gin.H{
			"user":    email,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	c.HTML(http.StatusOK, "thread.html", gin.H{
		"user":   email,
		"thread": thread,
		"posts":  thread.Posts,
	})
}

func (handler *ThreadHandler) NewThreadHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	fid := c.Param("fid")
	if c.Request.Method == "GET" {
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusOK, "new_thread.html", gin.H{
			"user": email,
			"fid":  fid,
			"csrf": csrf,
		})
		session.Save()
		return
	}

	db := models.DBConn
	newThread := new(models.NewThread)
	if err := c.Bind(&newThread); err != nil {
		session.AddFlash("Bad Request")
		c.HTML(http.StatusBadRequest, "new_thread.html", gin.H{
			"user": email,
			"fid":  fid,
			"csrf": fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}
	// Check forum exists
	var forum models.Forum
	db.Find(&forum, fid)
	if forum.ID == 0 {
		c.HTML(http.StatusNotFound, "new_thread.html", gin.H{
			"user": email,
			"fid":  fid,
			"csrf": fmt.Sprintf("%v", session.Get("csrf")),
		})
		return
	}
	user := getUserByEmail(email)
	forumId, _ := strconv.Atoi(fid)
	if strings.TrimSpace(newThread.Title) == "" || strings.TrimSpace(newThread.Body) == "" {
		session.AddFlash("Title and Body cannot be empty.")
		c.HTML(http.StatusBadRequest, "new_thread.html", gin.H{
			"user":    email,
			"fid":     fid,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	var thread models.Thread = models.Thread{
		Title:   newThread.Title,
		Body:    newThread.Body,
		ForumID: uint(forumId),
		Forum:   forum,
		UserID:  user.ID,
		User:    *user,
	}
	result := db.Create(&thread)
	if result.Error != nil {
		log.Printf("Error: %v", result.Error)
	}
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, fmt.Sprintf("forum%v", fid))
	c.Redirect(http.StatusFound, "/forums/"+fid)
}

// func DeleteThread(c *gin.Context)  {
// 	db := models.DBConn
// 	id := c.Param("id")
// 	var thread models.Thread
// 	db.Find(&thread, id)
// 	if thread.Title == "" {
// 		return c.Status(500).SendString("Thread does not exist in the database.")
// 	}

// 	db.Delete(&thread)
// 	return c.SendString("Thread successfully Deleted from database.")
// }

func (handler *ThreadHandler) UpdateThreadHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	id := c.Param("id")
	if c.Request.Method == "GET" {
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusOK, "new_thread.html", gin.H{
			"user": email,
			"id":   id,
			"csrf": csrf,
		})
		session.Save()
		return
	}
	db := models.DBConn
	var oldThread models.Thread
	var updThread = new(models.UpdateThread)
	db.First(&oldThread, id)
	if oldThread.ID == 0 {
		session.AddFlash("Invalid input")
		c.HTML(http.StatusBadRequest, "update_thread.html", gin.H{
			"user":    email,
			"id":      id,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		log.Printf("Invalid Forum ID. user: %s", email)
		return
	}

	// Check forum exists
	// fid := oldThread.ForumID
	// var forum models.Forum
	// db.Find(&forum, fid)
	// if forum.ID == 0 {
	// 	session.AddFlash("Invalid input")
	// 	c.HTML(http.StatusTeapot, "update_thread.html", gin.H{
	// 		"user":    email,
	// 		"id":      id,
	// 		"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
	// 		"flashes": session.Flashes(),
	// 	})
	// 	session.Save()
	// 	log.Printf("Invalid Forum ID. user: %f", email)
	// 	return
	// }

	// Get updated thread
	if err := c.Bind(updThread); err != nil {
		session.AddFlash("Invalid input")
		c.HTML(http.StatusBadRequest, "update_thread.html", gin.H{
			"user":    email,
			"id":      id,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	// Check CSRF
	if updThread.Csrf != fmt.Sprintf("%v", session.Get("csrf")) {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusBadRequest, "update_thread.html", gin.H{
			"user":    email,
			"id":      id,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	if strings.TrimSpace(updThread.Title) == "" || strings.TrimSpace(updThread.Body) == "" {
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		session.AddFlash("Invalid input")
		session.AddFlash("Title and Description cannot be empty.")
		c.HTML(http.StatusBadRequest, "update_thread.html", gin.H{
			"user":    email,
			"id":      id,
			"csrf":    csrf,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	// Update the thread/event
	oldThread.Title = updThread.Title
	oldThread.Body = updThread.Body
	oldThread.Date = updThread.Date

	db.Save(&oldThread)
	log.Println("==== Clearing Redis")
	redisClient := models.RedisClient
	redisClient.Del(c, fmt.Sprintf("forum%v", oldThread.ForumID))
	c.Redirect(http.StatusFound, "/thread/"+id)
}

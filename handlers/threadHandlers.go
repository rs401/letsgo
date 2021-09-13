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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		session.Save()
		return
	}
	// Check forum exists
	var forum models.Forum
	db.Find(&forum, fid)
	if forum.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Forum doesn't exist",
		})
		return
	}
	user := getUserByEmail(email)
	forumId, _ := strconv.Atoi(fid)
	if strings.TrimSpace(newThread.Title) == "" || strings.TrimSpace(newThread.Body) == "" {
		session.AddFlash("Title and Body cannot be empty.")
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusBadRequest, "new_thread.html", gin.H{
			"user":    email,
			"fid":     fid,
			"csrf":    csrf,
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

// func UpdateThread(c *gin.Context)  {
// 	db := models.DBConn
// 	id := c.Param("id")
// 	var oldThread models.Thread
// 	var updThread = new(models.Thread)
// 	db.First(&oldThread, id)
// 	// Check forum exists
// 	fid := oldThread.ForumID
// 	var forum models.Forum
// 	db.Find(&forum, fid)
// 	if forum.Name == "" {
// 		return c.Status(418).SendString("Forum doesn't exist")
// 	}

// 	if oldThread.Title == "" {
// 		return c.Status(500).SendString("Thread does not exist in the database.")
// 	}
// 	if err := c.BodyParser(updThread); err != nil {
// 		return c.Status(503).SendString(err.Error())
// 	}
// 	theId, idErr := strconv.Atoi(id)
// 	if idErr != nil {
// 		return c.Status(422).SendString(idErr.Error())
// 	}
// 	updThread.ID = uint(theId)
// 	if title := strings.TrimSpace(updThread.Title); title == "" {
// 		updThread.Title = oldThread.Title
// 	}
// 	if body := strings.TrimSpace(updThread.Body); body == "" {
// 		updThread.Body = oldThread.Body
// 	}
// 	if userid := updThread.UserID; userid == 0 {
// 		updThread.UserID = oldThread.UserID
// 	}
// 	if forumid := updThread.ForumID; forumid == 0 {
// 		updThread.ForumID = oldThread.ForumID
// 	}

// 	db.Save(&updThread)

// 	return c.JSON(updThread)
// }

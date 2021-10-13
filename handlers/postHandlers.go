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

// PostHandler handler func receiver
type PostHandler struct{}

// NewPostHandler returns the new post form on GET and creates new post on POST
func (handler *PostHandler) NewPostHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	tid := c.Param("tid")
	if c.Request.Method == "GET" {
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusOK, "new_post.gotmpl", gin.H{
			"user": email,
			"tid":  tid,
			"csrf": csrf,
		})
		session.Save()
		return
	}

	db := models.DBConn
	newPost := new(models.NewPost)
	if err := c.Bind(&newPost); err != nil {
		session.AddFlash("Bad Request")
		csrf := uuid.NewString()
		c.HTML(http.StatusBadRequest, "new_post.gotmpl", gin.H{
			"error":   err.Error(),
			"user":    email,
			"tid":     tid,
			"csrf":    csrf,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	// Check thread exists
	var thread models.Thread
	db.Find(&thread, tid)
	if thread.ID == 0 {
		session.AddFlash("Thread doesn't exist")
		csrf := uuid.NewString()
		c.HTML(http.StatusNotFound, "new_post.gotmpl", gin.H{
			"error":   "Thread doesn't exist",
			"user":    email,
			"tid":     tid,
			"csrf":    csrf,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	user := getUserByEmail(email)
	threadId, _ := strconv.Atoi(tid)
	if strings.TrimSpace(newPost.Body) == "" {
		session.AddFlash("Body cannot be empty.")
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusBadRequest, "new_post.gotmpl", gin.H{
			"user":    email,
			"tid":     tid,
			"csrf":    csrf,
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	var post models.Post = models.Post{
		Body:     newPost.Body,
		ThreadID: uint(threadId),
		Thread:   thread,
		UserID:   user.ID,
		User:     *user,
	}
	result := db.Create(&post)
	if result.Error != nil {
		log.Printf("Error: %v", result.Error)
	}
	c.Redirect(http.StatusFound, "/thread/"+tid)
}

// func DeletePost(c *gin.Context)  {
// 	db := models.DBConn
// 	id := c.Param("id")
// 	var post models.Post
// 	db.Find(&post, id)
// 	if post.Title == "" {
// 		return c.Status(500).SendString("Post does not exist in the database.")
// 	}

// 	db.Delete(&post)
// 	return c.SendString("Post successfully Deleted from database.")
// }

// func UpdatePost(c *gin.Context)  {
// 	db := models.DBConn
// 	id := c.Param("id")
// 	var oldPost models.Post
// 	var updPost = new(models.Post)
// 	db.First(&oldPost, id)
// 	// Check forum exists
// 	fid := oldPost.ThreadID
// 	var forum models.Thread
// 	db.Find(&forum, fid)
// 	if forum.Name == "" {
// 		return c.Status(418).SendString("Thread doesn't exist")
// 	}

// 	if oldPost.Title == "" {
// 		return c.Status(500).SendString("Post does not exist in the database.")
// 	}
// 	if err := c.BodyParser(updPost); err != nil {
// 		return c.Status(503).SendString(err.Error())
// 	}
// 	theId, idErr := strconv.Atoi(id)
// 	if idErr != nil {
// 		return c.Status(422).SendString(idErr.Error())
// 	}
// 	updPost.ID = uint(theId)
// 	if title := strings.TrimSpace(updPost.Title); title == "" {
// 		updPost.Title = oldPost.Title
// 	}
// 	if body := strings.TrimSpace(updPost.Body); body == "" {
// 		updPost.Body = oldPost.Body
// 	}
// 	if userid := updPost.UserID; userid == 0 {
// 		updPost.UserID = oldPost.UserID
// 	}
// 	if forumid := updPost.ThreadID; forumid == 0 {
// 		updPost.ThreadID = oldPost.ThreadID
// 	}

// 	db.Save(&updPost)

// 	return c.JSON(updPost)
// }

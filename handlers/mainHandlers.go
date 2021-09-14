package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs401/letsgo/models"
)

type MainHandler struct{}

func (handler *MainHandler) IndexHandler(c *gin.Context) {
	db := models.DBConn
	var forums []models.Forum
	db.Find(&forums)
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	if session.Get("email") == nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"forums": forums,
			"user":   email,
		})
		return
	}
	user := getUserByEmail(session.Get("email").(string))
	c.HTML(http.StatusOK, "index.html", gin.H{
		"forums": forums,
		"user":   user.Email,
	})
}

func (handler *MainHandler) ErrorHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.AddFlash(c.Errors.Errors())
	c.HTML(http.StatusOK, "error.html", gin.H{
		"flashes": c.Errors.Errors(),
	})
	session.Save()
}

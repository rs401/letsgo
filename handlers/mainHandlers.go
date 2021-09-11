package handlers

import (
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
	if session.Get("email") == nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"forums": forums,
		})
		return
	}
	user := getUserByEmail(session.Get("email").(string))
	c.HTML(http.StatusOK, "index.html", gin.H{
		"forums": forums,
		"user":   user.Email,
	})
}

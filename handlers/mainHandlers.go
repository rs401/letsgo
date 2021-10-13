package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs401/letsgo/models"
)

// MainHandler handler func receiver
type MainHandler struct{}

// IndexHandler returns index template
func (handler *MainHandler) IndexHandler(c *gin.Context) {
	db := models.DBConn
	var forums []models.Forum
	db.Find(&forums)
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	if session.Get("email") == nil {
		c.HTML(http.StatusOK, "index.gotmpl", gin.H{
			"forums": forums,
			"user":   email,
		})
		return
	}
	user := getUserByEmail(session.Get("email").(string))
	c.HTML(http.StatusOK, "index.gotmpl", gin.H{
		"forums": forums,
		"user":   user.Email,
	})
}

// PrivacyHandler returns privacy page
func (handler *MainHandler) PrivacyHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "privacy.gotmpl", gin.H{})
}

// TermsHandler returns terms of use page
func (handler *MainHandler) TermsHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "terms.gotmpl", gin.H{})
}

// ErrorHandler handles errors
func (handler *MainHandler) ErrorHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.AddFlash(c.Errors.Errors())
	c.HTML(http.StatusOK, "error.gotmpl", gin.H{
		"flashes": c.Errors.Errors(),
	})
	session.Save()
}

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs401/letsgo/models"
)

// UserHandler handler func receiver
type UserHandler struct{}

// GetUserHandler returns an update user form for account management
func (handler *UserHandler) GetUserHandler(c *gin.Context) {
	// get session and user
	session := sessions.Default(c)
	user := getUserByEmail(session.Get("email").(string))
	// CSRF
	csrf := uuid.NewString()
	session.Set("csrf", csrf)
	c.HTML(http.StatusOK, "account.html", gin.H{
		"csrf":    csrf,
		"user":    user.Email,
		"account": user,
	})
	session.Save()
}

// UpdateUserHandler updates a user account
func (handler *UserHandler) UpdateUserHandler(c *gin.Context) {
	// get session and user
	session := sessions.Default(c)
	user := getUserByEmail(session.Get("email").(string))
	// CSRF
	expectcsrf := fmt.Sprintf("%v", session.Get("csrf"))
	gotcsrf := c.Request.FormValue("csrf")
	if gotcsrf != expectcsrf {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusOK, "account.html", gin.H{
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
			"user":    fmt.Sprintf("%v", session.Get("email")),
			"account": user,
		})
		session.Save()
		return
	}
	// Update user
	if strings.TrimSpace(c.Request.FormValue("DisplayName")) == "" {
		session.AddFlash("Display Name cannot be empty.")
		c.HTML(http.StatusOK, "account.html", gin.H{
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
			"user":    fmt.Sprintf("%v", session.Get("email")),
			"account": user,
		})
		session.Save()
		return
	}
	if strings.TrimSpace(c.Request.FormValue("DisplayName")) != user.DisplayName {
		user.DisplayName = strings.TrimSpace(c.Request.FormValue("DisplayName"))
	}
	db := models.DBConn
	db.Save(&user)
	c.Redirect(http.StatusFound, "/")
}

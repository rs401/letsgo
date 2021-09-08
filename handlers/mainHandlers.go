package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs401/letsgo/models"
)

func IndexHandler(c *gin.Context) {
	db := models.DBConn
	var forums []models.Forum
	db.Find(&forums)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"forums": forums,
	})
}

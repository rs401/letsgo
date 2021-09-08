package handlers

import "github.com/gin-gonic/gin"

func IndexHandler(c *gin.Context) {
	// c.JSON(200, gin.H{
	// 	"message": "Let's Go",
	// })
	c.File("./static/index.html")
}

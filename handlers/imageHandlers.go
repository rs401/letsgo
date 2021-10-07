package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/disintegration/imaging"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs401/letsgo/models"
)

// ImageHandler handler func receiver
type ImageHandler struct{}

// UploadProfileImageHandler returns upload image template on GET and handles
// creating and resizing image on POST
func (handler *ImageHandler) UploadProfileImageHandler(c *gin.Context) {
	session := sessions.Default(c)
	email := fmt.Sprintf("%v", session.Get("email"))
	user := getUserByEmail(email)
	if c.Request.Method == "GET" {
		csrf := uuid.NewString()
		session.Set("csrf", csrf)
		c.HTML(http.StatusOK, "profile_image.gohtml", gin.H{
			"user": email,
			"csrf": csrf,
		})
		session.Save()
		return
	}
	recCsrf := c.Request.FormValue("csrf")
	// Check csrf
	if recCsrf != fmt.Sprintf("%v", session.Get("csrf")) {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusOK, "profile_image.gohtml", gin.H{
			"user":    email,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	db := models.DBConn
	uploadDir := "static/uploads/"
	fileName := uuid.NewString()
	filePath := uploadDir + fileName + ".jpg"
	file, err := c.FormFile("file")
	if err != nil {
		session.AddFlash("Invalid File")
		log.Printf("Error uploading image: %v", err)
		c.HTML(http.StatusBadRequest, "profile_image.gohtml", gin.H{
			"flashes": session.Flashes(),
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}

	uploadErr := c.SaveUploadedFile(file, filePath)
	if uploadErr != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("'%s' NOT uploaded!\n%v", file.Filename, uploadErr))
		return
	}
	src, err := imaging.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	dst := imaging.Resize(src, 300, 0, imaging.Lanczos)
	err = imaging.Save(dst, filePath)
	if err != nil {
		log.Printf("failed to save image: %v", err)
	}

	image := &models.Image{
		FileName: fileName,
		UserID:   user.ID,
		User:     *user,
	}
	db.Create(&image)
	user.Picture = image.FileName
	db.Save(&user)

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}

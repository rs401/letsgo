// Recipes API
//
// This is a sample recipes API. You can find out more about the API at https://github.com/rs401/letsgo/.
//
// Schemes: http
// Host: 127.0.0.1:9000
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"os"

	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs401/letsgo/handlers"
	"github.com/rs401/letsgo/models"
)

var mainHandler *handlers.MainHandler
var authHandler *handlers.AuthHandler
var forumHandler *handlers.ForumHandler
var threadHandler *handlers.ThreadHandler
var postHandler *handlers.PostHandler

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	models.InitDatabase()
	authHandler = &handlers.AuthHandler{}
	mainHandler = &handlers.MainHandler{}
	threadHandler = &handlers.ThreadHandler{}
	postHandler = &handlers.PostHandler{}

	// Seed some forums for testing
	// bytesRead, _ := ioutil.ReadFile("words.txt")
	// file_content := string(bytesRead)
	// words := strings.Split(file_content, "\n")
	// db := models.DBConn
	// for _, name := range words {
	// 	forum := models.Forum{Name: name}
	// 	db.Create(&forum)
	// }

}

func Config(key string) string {
	return os.Getenv(key)
}

func main() {
	r := gin.Default()

	if Config("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	store, _ := redisStore.NewStore(10, "tcp", Config("REDIS_HOST")+":"+Config("REDIS_PORT"), "", []byte(Config("SECRET")))
	store.Options(sessions.Options{MaxAge: 3600})
	r.Use(sessions.Sessions("letsgo_api", store))

	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./favicon.ico")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", mainHandler.IndexHandler)
	r.GET("/forums", forumHandler.GetForumsHandler)
	r.GET("/login", authHandler.LoginHandler)
	r.GET("/register", authHandler.RegisterHandler)
	r.POST("/register", authHandler.RegisterHandler)
	r.POST("/login", authHandler.LoginHandler)
	r.GET("/auth-callback", authHandler.CallbackHandler)
	r.POST("/refresh", authHandler.RefreshHandler)
	r.GET("/signout", authHandler.SignOutHandler)
	authorized := r.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.POST("/forums", forumHandler.NewForumHandler)
		authorized.GET("/new_forum", forumHandler.NewForumHandler)
		authorized.GET("/forums/:id", forumHandler.GetForumHandler)
		authorized.GET("/update_forum/:id", forumHandler.UpdateForumHandler)
		authorized.POST("/update_forum/:id", forumHandler.UpdateForumHandler)
		authorized.POST("/del_forum/:id", forumHandler.DeleteForumHandler)
		authorized.GET("/del_forum/:id", forumHandler.ConfirmDeleteForumHandler)

		authorized.GET("/new_thread/:fid", threadHandler.NewThreadHandler)
		authorized.POST("/new_thread/:fid", threadHandler.NewThreadHandler)
		authorized.GET("/thread/:id", threadHandler.GetThreadHandler)
		authorized.GET("/update_thread/:id", threadHandler.UpdateThreadHandler)
		authorized.POST("/update_thread/:id", threadHandler.UpdateThreadHandler)

		authorized.GET("/new_post/:tid", postHandler.NewPostHandler)
		authorized.POST("/new_post/:tid", postHandler.NewPostHandler)

	}

	r.RunTLS(":"+Config("API_PORT"), "./certs/localhost.crt", "./certs/localhost.key")
}

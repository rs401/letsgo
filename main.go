package main

import (
	"os"

	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/autotls"
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
var memberHandler *handlers.MemberHandler

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	models.InitDatabase()
	authHandler = &handlers.AuthHandler{}
	mainHandler = &handlers.MainHandler{}
	threadHandler = &handlers.ThreadHandler{}
	postHandler = &handlers.PostHandler{}
	memberHandler = &handlers.MemberHandler{}

}

func Config(key string) string {
	return os.Getenv(key)
}

func SetupServer() *gin.Engine {
	if Config("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

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

		authorized.POST("/request_membership/:id", memberHandler.RequestMembershipHandler)
		authorized.GET("/manage_members/:id", memberHandler.ManageMembersHandler)
		authorized.POST("/add_member/:id", memberHandler.AddMemberHandler)
		authorized.POST("/reject_member/:id", memberHandler.RejectMemberHandler)
		authorized.POST("/remove_member/:id", memberHandler.RemoveMemberHandler)

	}

	// r.RunTLS(":"+Config("API_PORT"), "./certs/localhost.crt", "./certs/localhost.key")

	return r
}

func main() {
	// SetupServer().RunTLS(":"+Config("API_PORT"), "./certs/localhost.crt", "./certs/localhost.key")
	autotls.Run(SetupServer(), "letsgo.events", "www.letsgo.events")
}

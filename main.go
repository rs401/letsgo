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

var authHandler *handlers.AuthHandler

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	models.InitDatabase()
	authHandler = &handlers.AuthHandler{}

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

	store, _ := redisStore.NewStore(10, "tcp", Config("REDIS_HOST")+":"+Config("REDIS_PORT"), "", []byte("secret"))
	store.Options(sessions.Options{MaxAge: 3600})
	r.Use(sessions.Sessions("letsgo_api", store))

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", handlers.IndexHandler)
	r.GET("/forums", handlers.GetForumsHandler)
	r.GET("/login", authHandler.LoginHandler)
	r.POST("/register", authHandler.RegisterHandler)
	r.POST("/signin", authHandler.SignInHandler)
	r.GET("/auth-callback", authHandler.CallbackHandler)
	r.POST("/refresh", authHandler.RefreshHandler)
	r.POST("/signout", authHandler.SignOutHandler)
	authorized := r.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.POST("/forums", handlers.NewForumHandler)
		authorized.GET("/forums/:id", handlers.GetForumHandler)
		authorized.PUT("/forums/:id", handlers.UpdateForumHandler)
		authorized.DELETE("/forums/:id", handlers.DeleteForumHandler)

	}
	// r.POST("/recipes", handlers.NewRecipeHandler)
	// r.GET("/recipes", handlers.ListRecipesHandler)
	// r.PUT("/recipes/:id", handlers.UpdateRecipeHandler)
	// r.DELETE("/recipes/:id", handlers.DeleteRecipeHandler)
	// r.GET("/recipes/search", handlers.SearchRecipesHandler)

	// r.Run(":" + Config("API_PORT"))
	r.RunTLS(":"+Config("API_PORT"), "./certs/localhost.crt", "./certs/localhost.key")
}

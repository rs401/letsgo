package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs401/letsgo/models"
	"golang.org/x/crypto/bcrypt"
)

var SecretKey = os.Getenv("JWT_SECRET")

type AuthHandler struct{}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func getUserByEmail(e string) *models.User {
	db := models.DBConn
	var user models.User
	db.Where("email = ?", e).Find(&user)
	if user.Email == e {
		return &user
	}
	return nil
}

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var auth0Domain = "https://" + os.Getenv("AUTH0_DOMAIN") + "/"
		// client := auth0.NewJWKClient(auth0.JWKClientOptions{
		// 	URI: auth0Domain + ".well-known/jwks.json",
		// }, nil)
		// configuration := auth0.NewConfiguration(client, []string{os.Getenv("AUTH0_API_IDENTIFIER")}, auth0Domain, jose.RS256)
		// validator := auth0.NewValidator(configuration, nil)
		// _, err := validator.ValidateRequest(c.Request)
		// if err != nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{
		// 		"message": "invalid token",
		// 	})
		// 	c.Abort()
		// 	return
		// }
		session := sessions.Default(c)
		sessionToken := session.Get("token")
		if sessionToken == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Not signed in",
			})
			c.Abort()
		}

		c.Next()
	}
}

func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	session := sessions.Default(c)
	sessionToken := session.Get("token")
	sessionUser := session.Get("email")
	if sessionToken == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session cookie"})
		return
	}

	sessionToken = uuid.NewString()
	session.Set("email", sessionUser.(string))
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "New session issued"})
}
func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var loginUser models.LoginUser

	if err := c.ShouldBindJSON(&loginUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
		})
		return
	}

	user := getUserByEmail(loginUser.Email)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "email/password incorrect",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(loginUser.Pass1)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "email/password incorrect",
		})
		return
	}

	sessionToken := uuid.NewString()
	session := sessions.Default(c)
	session.Set("email", user.Email)
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "User signed in"})
}

func (handler *AuthHandler) RegisterHandler(c *gin.Context) {
	var newUser models.NewUser

	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	if err := getUserByEmail(newUser.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email already exists in database",
		})
		return
	}
	// Check passwords match
	if newUser.Pass1 == "" || newUser.Pass1 != newUser.Pass2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Passwords do not match.",
		})
		return
	}
	password, err := bcrypt.GenerateFromPassword([]byte(newUser.Pass1), 14)
	if err != nil {
		fmt.Println("=========== Houston, we have a problem")
	}

	user := models.User{
		DisplayName: newUser.DisplayName,
		Email:       newUser.Email,
		Password:    password,
	}

	models.DBConn.Create(&user)

	c.JSON(http.StatusOK, user)
}

// swagger:operation POST /signout auth signOut
// Signing out
// ---
// responses:
//     '200':
//         description: Successful operation
func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/google/uuid"
	"github.com/rs401/letsgo/models"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var SecretKey = os.Getenv("JWT_SECRET")

type AuthHandler struct{}

type GUser struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Email   string `json:"email"`
}

var conf *oauth2.Config
var state string

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	conf = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "https://127.0.0.1:9000/auth-callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
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

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func (handler *AuthHandler) LoginHandler(c *gin.Context) {
	state = randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	url := getLoginURL(state)
	c.HTML(http.StatusOK, "login.html", gin.H{
		"gurl": url,
	})
}

func (handler *AuthHandler) CallbackHandler(c *gin.Context) {
	session := sessions.Default(c)
	if callbackState := session.Get("state"); callbackState != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: %s", callbackState))
		return
	}

	token, err := conf.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := conf.Client(context.Background(), token)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	fmt.Println("==== Email body: ", string(data))

	// read email, name, picture into GUser model
	var guser models.GUser
	if err := json.Unmarshal(data, &guser); err != nil {
		fmt.Println("Couldn't unmarshal data to guser: ", err.Error())
	}
	// check if user exists
	user := getUserByEmail(guser.Email)
	if user == nil {
		// if not exists, create new user
		user := models.User{
			DisplayName: guser.Name,
			Email:       guser.Email,
			Picture:     guser.Picture,
		}

		models.DBConn.Create(&user)
	}
	// if exists, set and update displayname and picture
	user.DisplayName = guser.Name
	user.Picture = guser.Picture
	models.DBConn.Save(&user)
	// Set auth session
	sessionToken := uuid.NewString()
	session = sessions.Default(c)
	session.Set("email", user.Email)
	session.Set("token", sessionToken)
	session.Save()

	c.Redirect(http.StatusTemporaryRedirect, "/")
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

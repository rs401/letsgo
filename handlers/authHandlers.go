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
			c.Redirect(http.StatusFound, "/login")
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

func (handler *AuthHandler) LoginHandler(c *gin.Context) {
	if c.Request.Method == "GET" {
		state = randToken()
		session := sessions.Default(c)
		session.Set("state", state)
		session.Save()
		url := getLoginURL(state)
		c.HTML(http.StatusOK, "login.html", gin.H{
			"gurl": url,
		})
		return
	}
	var loginUser models.LoginUser
	session := sessions.Default(c)

	if err := c.Bind(&loginUser); err != nil {
		session.AddFlash("Invalid Input")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error":   "Invalid input",
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	user := getUserByEmail(loginUser.Email)
	if user == nil {
		session.AddFlash("Email/Password Incorrect")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"message": "email/password incorrect",
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(loginUser.Pass1)); err != nil {
		session.AddFlash("Email/Password Incorrect")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"message": "email/password incorrect",
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	sessionToken := uuid.NewString()
	session.Set("email", user.Email)
	session.Set("token", sessionToken)
	session.AddFlash("User signed in")

	c.Redirect(http.StatusFound, "/")
	session.Save()
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
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
	session.AddFlash("User signed in")

	c.Redirect(http.StatusTemporaryRedirect, "/")
	session.Save()
}

func (handler *AuthHandler) RegisterHandler(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"message": "register",
		})
		return
	}
	var newUser models.NewUser
	session := sessions.Default(c)

	if err := c.Bind(&newUser); err != nil {
		session.AddFlash("invalid input")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"error":   "invalid input",
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	if err := getUserByEmail(newUser.Email); err != nil {
		session.AddFlash("email already exists in database")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"error":   "email already exists in database",
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	// Check passwords match
	if newUser.Pass1 == "" || newUser.Pass1 != newUser.Pass2 {
		session.AddFlash("Passwords do not match.")
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"error":   "Passwords do not match.",
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}
	password, err := bcrypt.GenerateFromPassword([]byte(newUser.Pass1), 14)
	if err != nil {
		fmt.Println("=========== Houston, we have a problem")
		session.AddFlash("Passwords do not match.")
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"error":   "internal server error",
			"flashes": session.Flashes(),
		})
		session.Save()
	}

	user := models.User{
		DisplayName: newUser.DisplayName,
		Email:       newUser.Email,
		Password:    password,
	}

	models.DBConn.Create(&user)

	c.Redirect(http.StatusFound, "login.html")
}

func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Save()
	session.Clear()
	c.Redirect(http.StatusFound, "/login")
}

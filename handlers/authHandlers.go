package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

// var secretKey = os.Getenv("JWT_SECRET")

// AuthHandler handler func receiver
type AuthHandler struct{}

var conf *oauth2.Config

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	conf = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GREDIRECT_URL"),
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

// AuthMiddleware Checks session for protected endpoints
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

// RefreshHandler refreshes session
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

// LoginHandler handles returning Login template for GET requests and handles
// login for POST requests
func (handler *AuthHandler) LoginHandler(c *gin.Context) {
	session := sessions.Default(c)
	var state, email string
	if c.Request.Method == "GET" {
		state = uuid.NewString()
		email = fmt.Sprintf("%v", session.Get("email"))
		url := getLoginURL(state)
		csrf := uuid.NewString()
		session.Set("state", state)
		session.Set("csrf", csrf)
		session.Save()
		c.HTML(http.StatusOK, "login.gotmpl", gin.H{
			"gurl": url,
			"user": email,
			"csrf": csrf,
		})
		return
	}
	var loginUser models.LoginUser

	if err := c.Bind(&loginUser); err != nil {
		session.AddFlash("Invalid Input")
		state = fmt.Sprintf("%v", session.Get("state"))
		url := getLoginURL(state)
		c.HTML(http.StatusBadRequest, "login.gotmpl", gin.H{
			"error":   "Invalid input",
			"flashes": session.Flashes(),
			"gurl":    url,
			"user":    email,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}

	user := getUserByEmail(loginUser.Email)
	if user == nil {
		session.AddFlash("Email/Password Incorrect")
		url := getLoginURL(state)
		c.HTML(http.StatusBadRequest, "login.gotmpl", gin.H{
			"message": "email/password incorrect",
			"flashes": session.Flashes(),
			"gurl":    url,
			"user":    email,
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}

	// Check CSRF
	if fmt.Sprintf("%v", session.Get("csrf")) != loginUser.Csrf {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusOK, "login.gotmpl", gin.H{
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(loginUser.Pass1)); err != nil {
		session.AddFlash("Email/Password Incorrect")
		c.HTML(http.StatusBadRequest, "login.gotmpl", gin.H{
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

// func randToken() string {
// 	b := make([]byte, 32)
// 	rand.Read(b)
// 	return base64.StdEncoding.EncodeToString(b)
// }

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

// CallbackHandler handles the exchange from Google Sign-in
func (handler *AuthHandler) CallbackHandler(c *gin.Context) {
	session := sessions.Default(c)
	callbackState := session.Get("state")
	if fmt.Sprintf("%v", callbackState) != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid session state: got: %v Type: %T\n  Expected: %v", callbackState, callbackState, c.Query("state")))
		return
	}

	token, err := conf.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := conf.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("==== Response body: ", string(data))

	// read email, name, picture into GUser model
	var guser models.GUser
	if err := json.Unmarshal(data, &guser); err != nil {
		fmt.Println("Couldn't unmarshal data to guser: ", err.Error())
	}
	// check if user exists
	var user *models.User
	user = getUserByEmail(guser.Email)
	if user == nil {
		// if not exists, create new user
		user = &models.User{
			DisplayName: guser.Name,
			Email:       guser.Email,
			Picture:     guser.Picture,
		}

		models.DBConn.Create(&user)
	}
	// if exists, set and update displayname and picture
	user.DisplayName = guser.Name
	if user.Picture == "" {
		user.Picture = guser.Picture
	}
	models.DBConn.Save(&user)
	// Set auth session
	sessionToken := uuid.NewString()
	session = sessions.Default(c)
	session.Set("email", user.Email)
	session.Set("token", sessionToken)
	session.AddFlash("User signed in")
	session.Save()

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// RegisterHandler handles returning register template for GET and registers
// users on POST
func (handler *AuthHandler) RegisterHandler(c *gin.Context) {
	var state string
	session := sessions.Default(c)
	if c.Request.Method == "GET" {
		state = uuid.NewString()
		url := getLoginURL(state)
		csrf := uuid.NewString()
		session.Set("state", state)
		session.Set("csrf", csrf)
		session.Save()
		c.HTML(http.StatusOK, "register.gotmpl", gin.H{
			"message": "register",
			"gurl":    url,
			"csrf":    csrf,
		})
		return
	}
	var newUser models.NewUser

	if err := c.Bind(&newUser); err != nil {
		session.AddFlash("invalid input")
		c.HTML(http.StatusBadRequest, "register.gotmpl", gin.H{
			"error":   "invalid input",
			"flashes": session.Flashes(),
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}

	if err := getUserByEmail(newUser.Email); err != nil {
		session.AddFlash("email already exists in database")
		c.HTML(http.StatusBadRequest, "register.gotmpl", gin.H{
			"error":   "email already exists in database",
			"flashes": session.Flashes(),
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}
	// Check passwords match
	if newUser.Pass1 == "" || newUser.Pass1 != newUser.Pass2 {
		session.AddFlash("Passwords do not match.")
		c.HTML(http.StatusBadRequest, "register.gotmpl", gin.H{
			"error":   "Passwords do not match.",
			"flashes": session.Flashes(),
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}
	// Check DisplayName not empty
	if newUser.DisplayName == "" {
		session.AddFlash("Display Name cannot be empty.")
		c.HTML(http.StatusBadRequest, "register.gotmpl", gin.H{
			"error":   "displayname empty",
			"flashes": session.Flashes(),
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
		})
		session.Save()
		return
	}

	// Check CSRF
	if fmt.Sprintf("%v", session.Get("csrf")) != newUser.Csrf {
		session.AddFlash("Cross Site Request Forgery")
		log.Println("==== CSRF did not match")
		log.Printf("==== %v", session.Get("email"))
		c.HTML(http.StatusOK, "register.gotmpl", gin.H{
			"csrf":    fmt.Sprintf("%v", session.Get("csrf")),
			"flashes": session.Flashes(),
		})
		session.Save()
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(newUser.Pass1), 14)
	if err != nil {
		fmt.Println("=========== Houston, we have a problem")
		session.AddFlash("Passwords do not match.")
		c.HTML(http.StatusInternalServerError, "register.gotmpl", gin.H{
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

	c.Redirect(http.StatusFound, "/login")
}

// SignOutHandler removes user session
func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("user")
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

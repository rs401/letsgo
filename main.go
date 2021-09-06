package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/xid"
)

type Recipe struct {
	ID           string    `json:"ID"`
	Name         string    `json:"Name"`
	Tags         []string  `json:"Tags"`
	Ingredients  []string  `json:"Ingredients"`
	Instructions []string  `json:"Instructions"`
	PublishedAt  time.Time `json:"PublishedAt"`
}

var recipes []Recipe

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("./recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}

func Config(key string) string {
	return os.Getenv(key)
}

func IndexHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Let's Go",
	})
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	recipe.ID = id
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found.",
		})
		return
	}
	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

func main() {
	port := Config("API_PORT")
	r := gin.Default()
	r.GET("/", IndexHandler)
	r.POST("/recipes", NewRecipeHandler)
	r.GET("/recipes", ListRecipesHandler)
	r.PUT("/recipes/:id", UpdateRecipeHandler)

	r.Run(":" + port)
}

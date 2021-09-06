package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

var recipes map[string]Recipe

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}
	var tmpRecipes []Recipe = make([]Recipe, 0)
	recipes = make(map[string]Recipe)
	file, _ := ioutil.ReadFile("./recipes.json")
	_ = json.Unmarshal([]byte(file), &tmpRecipes)
	for _, rec := range tmpRecipes {
		recipes[rec.ID] = rec
	}
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
	recipes[recipe.ID] = recipe
	// recipes = append(recipes, recipe)
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
	if tmpRecipe := recipes[id]; tmpRecipe.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}
	recipes[id] = recipe
	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	if tmpRecipe := recipes[id]; tmpRecipe.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}
	delete(recipes, id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe deleted",
	})
}

func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)
	for _, rec := range recipes {
		found := false
		for _, rtag := range rec.Tags {
			if strings.EqualFold(tag, rtag) {
				found = true
			}
		}
		if found {
			listOfRecipes = append(listOfRecipes, rec)
		}
	}
	c.JSON(http.StatusOK, listOfRecipes)
}

func main() {
	port := Config("API_PORT")
	r := gin.Default()
	r.GET("/", IndexHandler)
	r.POST("/recipes", NewRecipeHandler)
	r.GET("/recipes", ListRecipesHandler)
	r.PUT("/recipes/:id", UpdateRecipeHandler)
	r.DELETE("/recipes/:id", DeleteRecipeHandler)
	r.GET("/recipes/search", SearchRecipesHandler)

	r.Run(":" + port)
}

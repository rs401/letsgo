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

// NewRecipeHandler swagger:route POST /recipes recipes newRecipe
// Create new recipe
// ---
// produces:
// - application/json
// responses:
// 	'200':
// 		description: Successful operation
// 	'400':
// 		description: Invalid input
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

// ListRecipesHandler swagger:route GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
// 	'200':
// 		description: Successful operation
func ListRecipesHandler(c *gin.Context) {
	tmpList := make([]Recipe, 0)
	for _, rec := range recipes {
		tmpList = append(tmpList, rec)
	}
	c.JSON(http.StatusOK, tmpList)
}

// UpdateRecipeHandler swagger:route PUT /recipes/{id} updateRecipe
// Update an existing recipe
// ---
// parameters:
// 	- name: id
// 		in: path
// 		description: ID of the recipe
// 		required: true
// 		type: string
// 	- name: recipe
// 		in: body
// 		description: The updated recipe
// 		required: true
// 		type: json
// 		schema: Recipe
// produces:
// - application/json
// Schemes: http
// responses:
// 	'200':
// 		description: Successful operation
// 	'400':
// 		description: Invalid input
// 	'404':
// 		description: Invalid recipe ID
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

// DeleteRecipeHandler swagger:route DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe
// ---
// parameters:
// - name: id
// 	in: path
// 	description: ID of the recipe
// 	required: true
// 	type: string
// produces:
// - application/json
// responses:
// 	'200':
// 		description: Successful operation
// 	'404':
// 		description: Invalid recipe ID
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

// SearchRecipesHandler swagger:route PUT /recipes/search recipes searchRecipes
// Search existing recipes
// ---
// parameters:
// - name: tag
// 	in: query
// 	description: Tag to search for recipes
// 	required: true
// 	type: string
// produces:
// - application/json
// responses:
// 	'200':
// 		description: Successful operation
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

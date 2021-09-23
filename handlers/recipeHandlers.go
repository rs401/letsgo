package handlers

// import (
// 	"net/http"
// 	"strconv"

// 	// "strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/rs401/letsgo/models"
// )

// // NewRecipeHandler swagger:route POST /recipes recipes newRecipe
// // Create new recipe
// // ---
// // produces:
// // - application/json
// // responses:
// // 	'200':
// // 		description: Successful operation
// // 	'400':
// // 		description: Invalid input
// func NewRecipeHandler(c *gin.Context) {
// 	var recipe models.Recipe
// 	if err := c.ShouldBindJSON(&recipe); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}

// 	recipe.PublishedAt = time.Now()
// 	models.DBConn.Create(&recipe)
// 	c.JSON(http.StatusOK, recipe)
// }

// // ListRecipesHandler swagger:route GET /recipes recipes listRecipes
// // Returns list of recipes
// // ---
// // produces:
// // - application/json
// // responses:
// // 	'200':
// // 		description: Successful operation
// func ListRecipesHandler(c *gin.Context) {
// 	var tmpList []models.Recipe
// 	if result := models.DBConn.Find(&tmpList); result.Error != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": result.Error.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, tmpList)
// }

// // swagger:operation PUT /recipes/{id} recipes updateRecipe
// // "Update an existing recipe"
// // --
// // parameters:
// // 	- name: id
// // 		in: path
// // 		description: ID of the recipe
// // 		required: true
// // 		type: string
// //
// // produces:
// // 	- application/json
// //
// // responses:
// // 	'200':
// // 		description: "Successful operation"
// // 	'400':
// // 		description: "Invalid input"
// // 	'404':
// // 		description: "Invalid recipe ID"
// func UpdateRecipeHandler(c *gin.Context) {
// 	id, err := strconv.Atoi(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Bad request",
// 		})
// 		return
// 	}
// 	var recipe models.Recipe
// 	if err := c.ShouldBindJSON(&recipe); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}
// 	var tmpRecipe models.Recipe
// 	models.DBConn.First(&tmpRecipe, id)
// 	if tmpRecipe.ID == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error": "Recipe not found",
// 		})
// 		return
// 	}
// 	// merge recipe into tmpRecipe
// 	models.DBConn.Model(&tmpRecipe).Updates(recipe)
// 	c.JSON(http.StatusOK, tmpRecipe)
// }

// // DeleteRecipeHandler swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// // Delete an existing recipe
// // ---
// // parameters:
// // - name: id
// // 	in: path
// // 	description: ID of the recipe
// // 	required: true
// // 	type: string
// // produces:
// // - application/json
// // responses:
// // 	'200':
// // 		description: Successful operation
// // 	'404':
// // 		description: Invalid recipe ID
// func DeleteRecipeHandler(c *gin.Context) {
// 	id := c.Param("id")
// 	var tmpRecipe models.Recipe
// 	res := models.DBConn.Find(&tmpRecipe, id)
// 	if res.Error != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "Forum does not exist in the database.",
// 		})
// 		return
// 	}
// 	if tmpRecipe.ID == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error": "Recipe not found",
// 		})
// 		return
// 	}

// 	models.DBConn.Delete(&tmpRecipe)
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Recipe deleted",
// 	})
// }

// // SearchRecipesHandler swagger:operation PUT /recipes/search recipes searchRecipes
// // Search existing recipes
// // ---
// // parameters:
// // - name: tag
// // 	in: query
// // 	description: Tag to search for recipes
// // 	required: true
// // 	type: string
// // produces:
// // - application/json
// // responses:
// // 	'200':
// // 		description: Successful operation
// func SearchRecipesHandler(c *gin.Context) {
// 	tag := c.Query("tag")
// 	listOfRecipes := make([]models.Recipe, 0)
// 	var recipes []models.Recipe
// 	models.DBConn.Where("tags <> ?", tag).Find(&recipes)
// 	// for _, rec := range recipes {
// 	// 	found := false
// 	// 	for _, rtag := range rec.Tags {
// 	// 		if strings.EqualFold(tag, rtag) {
// 	// 			found = true
// 	// 		}
// 	// 	}
// 	// 	if found {
// 	// 		listOfRecipes = append(listOfRecipes, rec)
// 	// 	}
// 	// }
// 	c.JSON(http.StatusOK, listOfRecipes)
// }

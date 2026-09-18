package drawref

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// uri params

type CategoryRequest struct {
	Slug string `uri:"slug" binding:"required"`
}

// request bodies

type CreateCategoryParams struct {
	ID          string          `json:"id" binding:"required"`
	DisplayName *string         `json:"display_name"`
	CoverImage  int             `json:"cover_image"`
	Tags        json.RawMessage `json:"tags"`
	Position    int             `json:"position"`
}

type EditCategoryParams struct {
	DisplayName *string         `json:"display_name"`
	CoverImage  int             `json:"cover_image"`
	Tags        json.RawMessage `json:"tags"`
	Position    int             `json:"position"`
}

type ReorderCategoriesParams struct {
	IDs []string `json:"ids" binding:"required"`
}

// handlers

func getCategories(c *gin.Context) {
	categories, err := TheDb.GetCategories()
	if err != nil {
		fmt.Println("Could not get categories:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch categories"})
		return
	}

	// return an empty array [] when there are no categories
	if categories == nil {
		categories = []Category{}
	}

	c.JSON(http.StatusOK, categories)
}

func getCategory(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category slug must be provided"})
		return
	}

	category, err := TheDb.GetCategory(req.Slug)
	if err != nil {
		fmt.Println("Could not get category:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func createCategory(c *gin.Context) {
	var params CreateCategoryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// default tags to empty array
	if len(params.Tags) == 0 {
		params.Tags = json.RawMessage("[]")
	}

	newCategory := &Category{
		ID:          params.ID,
		DisplayName: params.DisplayName,
		CoverImage:  params.CoverImage,
		Tags:        params.Tags,
		Position:    params.Position,
	}

	err := TheDb.CreateCategory(newCategory)
	if err != nil {
		fmt.Println("Could not create category:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't create category. Maybe it already exists?"})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "create_category", "Created category "+params.ID, map[string]string{
		"id": params.ID,
	})

	c.JSON(http.StatusOK, gin.H{"id": params.ID})
}

func editCategory(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category slug must be provided"})
		return
	}

	var params EditCategoryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(params.Tags) == 0 {
		params.Tags = json.RawMessage("[]")
	}

	updateCategory := &Category{
		DisplayName: params.DisplayName,
		CoverImage:  params.CoverImage,
		Tags:        params.Tags,
		Position:    params.Position,
	}

	err := TheDb.UpdateCategory(req.Slug, updateCategory)
	if err != nil {
		fmt.Println("Could not edit category:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't edit category"})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "edit_category", "Updated category "+req.Slug, map[string]string{
		"id": req.Slug,
	})

	c.JSON(http.StatusOK, gin.H{"id": req.Slug})
}

func deleteCategory(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category slug must be provided"})
		return
	}

	err := TheDb.DeleteCategory(req.Slug)
	if err != nil {
		fmt.Println("Could not delete category:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't delete category"})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "delete_category", "Deleted category "+req.Slug, map[string]string{
		"id": req.Slug,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func reorderCategories(c *gin.Context) {
	var params ReorderCategoriesParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := TheDb.ReorderCategories(params.IDs)
	if err != nil {
		fmt.Println("Could not reorder categories:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't reorder categories"})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "reorder_categories", "Reordered categories", map[string]interface{}{
		"count": len(params.IDs),
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

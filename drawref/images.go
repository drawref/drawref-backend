package drawref

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// uri params

type ImageRequest struct {
	ID int `uri:"id" binding:"required"`
}

// request bodies

type UpdateImageParams struct {
	CategoryOverride  *string         `json:"category_override"`
	AuthorOverride    *string         `json:"author_override"`
	AuthorURLOverride *string         `json:"author_url_override"`
	TagsOverride      json.RawMessage `json:"tags_override"`
}

// handlers

func serveImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	img, err := TheDb.GetImage(req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	source, err := TheDb.GetSource(img.SourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch image source"})
		return
	}

	if source.SourceType == "local" {
		fullPath := filepath.Join(source.RootPath, img.RelativePath)
		// c.File automatically handles setting the correct Content-Type
		c.File(fullPath)
		return
	}

	if source.SourceType == "s3" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "S3 serving not implemented or URL is missing"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown source type"})
}

func getImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	img, err := TheDb.GetImage(req.ID)
	if err != nil {
		fmt.Println("Could not get image:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, img)
}

func updateImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	var params UpdateImageParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// load base image to ensure existence and keep SourceID etc intact
	img, err := TheDb.GetImage(req.ID)
	if err != nil {
		fmt.Println("Could not find image to update:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// apply any given overrides
	img.CategoryOverride = params.CategoryOverride
	img.AuthorOverride = params.AuthorOverride
	img.AuthorURLOverride = params.AuthorURLOverride
	img.TagsOverride = params.TagsOverride

	// database update triggers metadata recalculation automatically
	err = TheDb.UpdateImage(img)
	if err != nil {
		fmt.Println("Could not update image:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't update image", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "update_image", fmt.Sprintf("Updated image overrides for image %d", req.ID), map[string]interface{}{
		"id": req.ID,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func deleteImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	err := TheDb.DeleteImage(req.ID)
	if err != nil {
		fmt.Println("Could not delete image:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't delete image", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "delete_image", fmt.Sprintf("Deleted image %d", req.ID), map[string]interface{}{
		"id": req.ID,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

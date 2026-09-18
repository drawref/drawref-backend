package drawref

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// uri params

type SourceMetadataRequest struct {
	SourceID int `uri:"slug" binding:"required"`
}

// request bodies

type UpsertPathMetadataParams struct {
	RelativePath string          `json:"relative_path"` // empty string for root
	CategoryID   *string         `json:"category_id"`
	Author       *string         `json:"author"`
	AuthorURL    *string         `json:"author_url"`
	Tags         json.RawMessage `json:"tags"`
	TagMode      string          `json:"tag_mode"` // "merge" or "replace"
}

type DeletePathMetadataParams struct {
	ID int `json:"id" binding:"required"`
}

// handlers

func getSourcePathMetadata(c *gin.Context) {
	var req SourceMetadataRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Source ID must be provided"})
		return
	}

	metadata, err := TheDb.GetPathMetadataBySource(req.SourceID)
	if err != nil {
		fmt.Println("Could not get source path metadata:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch path metadata", "details": err.Error()})
		return
	}

	if metadata == nil {
		metadata = []PathMetadata{}
	}

	c.JSON(http.StatusOK, metadata)
}

func upsertPathMetadata(c *gin.Context) {
	var req SourceMetadataRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Source ID must be provided"})
		return
	}

	var params UpsertPathMetadataParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default values and null checks
	if len(params.Tags) == 0 {
		params.Tags = json.RawMessage("{}") // Note: Database specifies '{}' default for tags, not '[]'
	}
	if params.TagMode == "" {
		params.TagMode = "merge"
	}

	fm := &PathMetadata{
		SourceID:     req.SourceID,
		RelativePath: params.RelativePath,
		CategoryID:   params.CategoryID,
		Author:       params.Author,
		AuthorURL:    params.AuthorURL,
		Tags:         params.Tags,
		TagMode:      params.TagMode,
	}

	err := TheDb.UpsertPathMetadata(fm)
	if err != nil {
		fmt.Println("Could not upsert path metadata:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't save path metadata", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelDebug, "upsert_metadata", fmt.Sprintf("Updated metadata on source %d path '%s'", req.SourceID, params.RelativePath), map[string]interface{}{
		"source_id":     req.SourceID,
		"relative_path": params.RelativePath,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func deletePathMetadata(c *gin.Context) {
	var req SourceMetadataRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Source ID must be provided"})
		return
	}

	var params DeletePathMetadataParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := TheDb.DeletePathMetadata(params.ID)
	if err != nil {
		fmt.Println("Could not delete path metadata:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't delete path metadata", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelDebug, "delete_metadata", fmt.Sprintf("Deleted metadata entry [%d] on source [%d]", params.ID, req.SourceID), map[string]interface{}{
		"source_id": req.SourceID,
		"id":        params.ID,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

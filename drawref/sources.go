package drawref

import (
	"fmt"
	"net/http"

	"github.com/drawref/drawref-backend/drawref/sources"
	"github.com/gin-gonic/gin"
)

// uri params

type SourceRequest struct {
	ID int `uri:"slug" binding:"required"`
}

// request bodies

type CreateSourceParams struct {
	Name       string `json:"name" binding:"required"`
	SourceType string `json:"source_type" binding:"required"`
	RootPath   string `json:"root_path" binding:"required"`
	Enabled    bool   `json:"enabled"`
}

type EditSourceParams struct {
	Name       string `json:"name" binding:"required"`
	SourceType string `json:"source_type" binding:"required"`
	RootPath   string `json:"root_path" binding:"required"`
	Enabled    bool   `json:"enabled"`
}

// handlers

func getSources(c *gin.Context) {
	sources, err := TheDb.GetSources()
	if err != nil {
		fmt.Println("Could not get sources:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch sources"})
		return
	}

	if sources == nil {
		sources = []Source{}
	}

	c.JSON(http.StatusOK, sources)
}

func getSource(c *gin.Context) {
	var req SourceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Source ID must be provided"})
		return
	}

	source, err := TheDb.GetSource(req.ID)
	if err != nil {
		fmt.Println("Could not get source:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	c.JSON(http.StatusOK, source)
}

func createSource(c *gin.Context) {
	var params CreateSourceParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newSource := &Source{
		Name:       params.Name,
		SourceType: params.SourceType,
		RootPath:   params.RootPath,
		Enabled:    params.Enabled,
	}

	err := TheDb.CreateSource(newSource)
	if err != nil {
		fmt.Println("Could not create source:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't create source. Maybe one already exists for this type and root path?", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "create_source", "Created source: "+params.Name, map[string]interface{}{
		"id":        newSource.ID,
		"root_path": newSource.RootPath,
	})

	c.JSON(http.StatusOK, gin.H{"id": newSource.ID})
}

func editSource(c *gin.Context) {
	var req SourceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Source ID must be provided"})
		return
	}

	var params EditSourceParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateSource := &Source{
		ID:         req.ID,
		Name:       params.Name,
		SourceType: params.SourceType,
		RootPath:   params.RootPath,
		Enabled:    params.Enabled,
	}

	err := TheDb.UpdateSource(updateSource)
	if err != nil {
		fmt.Println("Could not edit source:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't edit source", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "edit_source", fmt.Sprintf("Updated source %d", req.ID), map[string]interface{}{
		"id": req.ID,
	})

	c.JSON(http.StatusOK, gin.H{"id": req.ID})
}

func deleteSource(c *gin.Context) {
	var req SourceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Source ID must be provided"})
		return
	}

	err := TheDb.DeleteSource(req.ID)
	if err != nil {
		fmt.Println("Could not delete source:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't delete source", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "delete_source", fmt.Sprintf("Deleted source %d", req.ID), map[string]interface{}{
		"id": req.ID,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func scanSource(c *gin.Context) {
	var req SourceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Source ID must be provided"})
		return
	}

	source, err := TheDb.GetSource(req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "scan_source", fmt.Sprintf("Triggered scan for source %d", req.ID), map[string]interface{}{
		"id": req.ID,
	})

	err = sources.ScanLocalSource(c.Request.Context(), source.ID, source.RootPath, TheDb)
	if err != nil {
		fmt.Println("Could not process scan:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process scan for source", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Source scan triggered"})
}

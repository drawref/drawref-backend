package drawref

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// uri + query params

type SessionQueryRequest struct {
	Category string `form:"category" binding:"required"`
	Tags     string `form:"tags"`
	Limit    int    `form:"limit"`
}

// handlers

func getSession(c *gin.Context) {
	var req SessionQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required parameters (category) not supplied or invalid"})
		return
	}

	// Parse tags
	var tags json.RawMessage
	if req.Tags == "" {
		tags = json.RawMessage("{}")
	} else {
		if !json.Valid([]byte(req.Tags)) {
			fmt.Println("Failed to parse tags:", req.Tags)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't parse tags"})
			return
		}
		tags = json.RawMessage(req.Tags)
	}

	// set default limit if none provided
	limit := req.Limit
	if limit <= 0 {
		limit = 30
	}

	images, err := TheDb.GetSessionImages(req.Category, tags, limit)
	if err != nil {
		fmt.Println("Could not get session images:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch session images"})
		return
	}

	// return an empty array instead of null
	if images == nil {
		images = []Image{}
	}

	c.JSON(http.StatusOK, images)
}

func getSessionCount(c *gin.Context) {
	var req SessionQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required parameters (category) not supplied or invalid."})
		return
	}

	// parse tags
	var tags json.RawMessage
	if req.Tags == "" {
		tags = json.RawMessage("{}")
	} else {
		if !json.Valid([]byte(req.Tags)) {
			fmt.Println("Failed to parse tags:", req.Tags)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't parse tags"})
			return
		}
		tags = json.RawMessage(req.Tags)
	}

	count, err := TheDb.GetSessionImageCount(req.Category, tags)
	if err != nil {
		fmt.Println("Could not get session image count:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch session image count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": count,
	})
}

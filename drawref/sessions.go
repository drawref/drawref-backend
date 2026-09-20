package drawref

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// uri + query params

type SessionQueryRequest struct {
	Category string `form:"category" binding:"required"`
	Tags     string `form:"tags"`
	Limit    int    `form:"limit"`
	Source   string `form:"source"`
}

// strips any keys with empty arrays or nil values, so thzt
// searching works as expected
func CleanSessionTags(raw json.RawMessage) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || trimmed == "{}" || trimmed == "[]" {
		return json.RawMessage("{}"), nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return raw, err
	}

	cleaned := make(map[string]interface{})
	for k, v := range parsed {
		if v == nil {
			continue
		}
		if arr, ok := v.([]interface{}); ok {
			if len(arr) == 0 {
				continue
			}
		}
		cleaned[k] = v
	}

	bytes, err := json.Marshal(cleaned)
	if err != nil {
		return raw, err
	}
	return json.RawMessage(bytes), nil
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
		cleaned, err := CleanSessionTags(json.RawMessage(req.Tags))
		if err != nil {
			fmt.Println("Failed to parse tags:", req.Tags, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't parse tags"})
			return
		}
		tags = cleaned
	}

	// Parse source
	var sourceID *int
	if req.Source != "" {
		id, err := strconv.Atoi(req.Source)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't parse source"})
			return
		}
		sourceID = &id
	}

	// set default limit if none provided
	limit := req.Limit
	if limit <= 0 {
		limit = 30
	}

	images, err := TheDb.GetSessionImages(req.Category, tags, sourceID, limit)
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
		cleaned, err := CleanSessionTags(json.RawMessage(req.Tags))
		if err != nil {
			fmt.Println("Failed to parse tags:", req.Tags, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't parse tags"})
			return
		}
		tags = cleaned
	}

	// Parse source
	var sourceID *int
	if req.Source != "" {
		id, err := strconv.Atoi(req.Source)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't parse source"})
			return
		}
		sourceID = &id
	}

	count, err := TheDb.GetSessionImageCount(req.Category, tags, sourceID)
	if err != nil {
		fmt.Println("Could not get session image count:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch session image count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": count,
	})
}

package drawref

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func autocompleteDirectory(c *gin.Context) {
	queryPath := c.Query("path")

	// provide a default root if empty
	if queryPath == "" {
		queryPath = string(filepath.Separator)
	}

	dir, prefix := filepath.Split(queryPath)

	// clean the dir path to handle weird formatting
	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		// return an empty array if the folder doesn't exist or permission is denied
		c.JSON(http.StatusOK, []string{})
		return
	}

	var suggestions []string
	for _, entry := range entries {
		// only suggest directories and match the prefix
		if entry.IsDir() && strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(prefix)) {
			fullPath := filepath.Join(dir, entry.Name()) + string(filepath.Separator)
			suggestions = append(suggestions, fullPath)
		}
	}

	c.JSON(http.StatusOK, suggestions)
}

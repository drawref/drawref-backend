package drawref

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Logs a failed login attempt.
func LogFailedLoginAttempt(address string) error {
	return Log(LogLevelInfo, "login-failed", "Failed login attempt", map[string]string{
		"address": address,
	})
}

// Logs a successful login attempt.
func LogSuccessfulLoginAttempt(level string) error {
	return Log(LogLevelInfo, "login-success", "Successful login attempt for "+level, map[string]string{
		"level": level,
	})
}

// Logs a new line.
func Log(logLevel LogLevel, eventType string, message string, data interface{}) error {
	return TheDb.AddLogLine(logLevel, eventType, message, data)
}

// handler

type LogsRequest struct {
	Page int `form:"page"`
}

type LogsResponse struct {
	Logs []LogLine `json:"logs"`
}

func getLogs(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return
	}

	response, err := TheDb.GetLogs(page)
	if err != nil {
		fmt.Println("Could not get logs:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get logs"})
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"logs": response,
	})
}

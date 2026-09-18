package drawref

import (
	"time"
)

// logs

type LogLevel string

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
)

type LogLine struct {
	ID        int         `json:"id"`
	Time      time.Time   `json:"ts"`
	Level     LogLevel    `json:"level"`
	EventType string      `json:"event"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

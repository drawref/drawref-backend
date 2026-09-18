package drawref

import (
	"encoding/json"
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

// categories

type Category struct {
	ID          string          `json:"id"`
	DisplayName *string         `json:"display_name"`
	CoverImage  int             `json:"cover_image"`
	Tags        json.RawMessage `json:"tags"`
	Position    int             `json:"position"`
}

// sources

type Source struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	SourceType    string     `json:"source_type"`
	RootPath      string     `json:"root_path"`
	Enabled       bool       `json:"enabled"`
	LastScannedAt *time.Time `json:"last_scanned_at"`
}

type PathMetadata struct {
	ID           int             `json:"id"`
	SourceID     int             `json:"source_id"`
	RelativePath string          `json:"relative_path"`
	CategoryID   *string         `json:"category_id"`
	Author       *string         `json:"author"`
	AuthorURL    *string         `json:"author_url"`
	Tags         json.RawMessage `json:"tags"`
	TagMode      string          `json:"tag_mode"`
}

// images

type Image struct {
	ID                  int             `json:"id"`
	SourceID            int             `json:"source_id"`
	RelativePath        string          `json:"relative_path"`
	FileHash            *string         `json:"file_hash"`
	LocalPath           *string         `json:"local_path"`
	ExternalURL         *string         `json:"external_url"`
	CategoryOverride    *string         `json:"category_override"`
	AuthorOverride      *string         `json:"author_override"`
	AuthorURLOverride   *string         `json:"author_url_override"`
	TagsOverride        json.RawMessage `json:"tags_override"`
	EffectiveCategoryID *string         `json:"effective_category_id"`
	EffectiveAuthor     *string         `json:"effective_author"`
	EffectiveAuthorURL  *string         `json:"effective_author_url"`
	EffectiveTags       json.RawMessage `json:"effective_tags"`
}

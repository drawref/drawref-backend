package drawref

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/drawref/drawref-backend/drawref/sources"
	"github.com/gin-gonic/gin"
)

//go:embed samples.json
var samplesJSON []byte

//go:embed sample-images
var SamplesFS embed.FS

type SampleData struct {
	Categories []SampleCategory `json:"categories"`
	Providers  []SampleProvider `json:"providers"`
}
type SampleCategory struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Cover   string `json:"cover,omitempty"`
	TagsRaw string `json:"tags_raw"`
}
type SampleProvider struct {
	Author      string             `json:"author"`
	AuthorURL   string             `json:"author_url"`
	Collections []SampleCollection `json:"collections"`
}
type SampleCollection struct {
	Category string        `json:"category"`
	Images   []SampleImage `json:"images"`
}
type SampleImage struct {
	Path string   `json:"path"`
	Tags []string `json:"tags"`
}

func loadSamples(c *gin.Context) {
	var data SampleData
	if err := json.Unmarshal(samplesJSON, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse samples.json"})
		return
	}

	ctx := context.Background()

	// fetch or create the sample source
	var sourceID int
	err := TheDb.pool.QueryRow(ctx, "SELECT id FROM sources WHERE root_path = $1", "internal://samples").Scan(&sourceID)
	if err != nil {
		source := &Source{
			Name:       "Sample Data",
			SourceType: "samples",
			RootPath:   "internal://samples",
			Enabled:    true,
		}
		if err = TheDb.CreateSource(source); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create source"})
			return
		}
		sourceID = source.ID
	}

	// create sample categories with default cover image
	for i, cat := range data.Categories {
		tagsList := parseSampleCategoryTags(cat.TagsRaw)
		tagsJSON, _ := json.Marshal(tagsList)

		_, err := TheDb.pool.Exec(ctx, `
            INSERT INTO categories (id, display_name, position, tags) 
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (id) DO NOTHING`,
			cat.ID, cat.Name, (i+1)*100, tagsJSON)
		if err != nil {
			fmt.Println("Error inserting category:", err)
		}
	}

	// insert images with overrides
	for _, prov := range data.Providers {
		for _, col := range prov.Collections {
			for _, img := range col.Images {
				tagsMap := parseSampleTags(img.Tags)
				tagsJSON, _ := json.Marshal(tagsMap)

				_, err := TheDb.pool.Exec(ctx, `
                    INSERT INTO images (source_id, relative_path, category_override, author_override, author_url_override, tags_override)
                    VALUES ($1, $2, $3, $4, $5, $6)
                    ON CONFLICT (source_id, relative_path) DO UPDATE SET
                        category_override = EXCLUDED.category_override,
                        author_override = EXCLUDED.author_override,
                        author_url_override = EXCLUDED.author_url_override,
                        tags_override = EXCLUDED.tags_override
                    `, sourceID, img.Path, col.Category, prov.Author, prov.AuthorURL, tagsJSON)
				if err != nil {
					fmt.Println("Error inserting image override:", err)
				}
			}
		}
	}

	// don't run in the background, we need images to exist to set covers on new categories
	err = sources.ScanFSSource(context.Background(), sourceID, SamplesFS, "sample-images", TheDb)
	if err != nil {
		fmt.Println("Error scanning sample images:", err)
	}

	// set cover images
	for _, cat := range data.Categories {
		if cat.Cover == "" {
			continue
		}

		// look up the image ID that matches the cover path
		var imageID int
		err := TheDb.pool.QueryRow(ctx, `
            SELECT id FROM images 
            WHERE source_id = $1 AND relative_path = $2
        `, sourceID, cat.Cover).Scan(&imageID)

		if err == nil {
			// link the image ID back to the category
			_, err = TheDb.pool.Exec(ctx, `
                UPDATE categories SET cover_image = $1 WHERE id = $2
            `, imageID, cat.ID)
			if err != nil {
				fmt.Println("Error updating category cover:", err)
			}
		} else {
			fmt.Printf("Could not find cover image '%s' for category '%s': %v\n", cat.Cover, cat.ID, err)
		}
	}

	TheDb.AddLogLine(LogLevelInfo, "load_samples", "Sample data loaded and scan started", nil)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// tag helpers mirroring the frontend parsers

func parseSampleTags(input []string) map[string][]string {
	tags := make(map[string][]string)
	for _, group := range input {
		parts := strings.SplitN(group, " ", 2)
		if len(parts) > 0 {
			tId := parts[0]
			val := ""
			if len(parts) > 1 {
				val = strings.TrimSpace(parts[1])
			}
			tags[tId] = append(tags[tId], val)
		}
	}
	return tags
}

func parseSampleCategoryTags(input string) []map[string]interface{} {
	var result []map[string]interface{}
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			id := strings.ReplaceAll(strings.ToLower(name), " ", "_")
			var values []string
			for _, v := range strings.Split(parts[1], ",") {
				values = append(values, strings.TrimSpace(v))
			}
			result = append(result, map[string]interface{}{
				"id": id, "name": name, "values": values,
			})
		}
	}
	if result == nil {
		return make([]map[string]interface{}, 0)
	}
	return result
}

package drawref

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultLogLinesPerPage = 150

type DRDatabase struct {
	pool *pgxpool.Pool
}

var TheDb *DRDatabase

func OpenDatabase(connectionUrl string) error {
	pool, err := pgxpool.New(context.Background(), connectionUrl)
	if err != nil {
		return err
	}

	TheDb = &DRDatabase{
		pool,
	}

	return nil
}

func (db *DRDatabase) Close() {
	db.pool.Close()
}

// log lines

func (db *DRDatabase) AddLogLine(logLevel LogLevel, eventType string, message string, data interface{}) error {
	fmt.Println("Log:", logLevel, eventType, message, data)

	_, err := db.pool.Exec(context.Background(), `
insert into logs (log_level, event_type, message, extra_data)
values ($1, $2, $3, $4)
	`, logLevel, eventType, message, data)

	return err
}

func (db *DRDatabase) GetLogs(page int) (logs []LogLine, err error) {
	startLog := max(page, 0) * DefaultLogLinesPerPage

	rows, err := db.pool.Query(context.Background(), `
SELECT id, ts, log_level, event_type, message, extra_data
FROM logs
ORDER BY ts desc
LIMIT $1
OFFSET $2
`, DefaultLogLinesPerPage, startLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Log Query failed: %v\n", err)
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		var l LogLine

		err = rows.Scan(&l.ID, &l.Time, &l.Level, &l.EventType, &l.Message, &l.Data)
		if err != nil {
			return logs, err
		}
		logs = append(logs, l)
	}

	return logs, err
}

// categories

func (db *DRDatabase) GetCategories() ([]Category, error) {
	rows, err := db.pool.Query(context.Background(), `
        SELECT id, display_name, cover_image, tags, position
        FROM categories
        ORDER BY position ASC, id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.DisplayName, &c.CoverImage, &c.Tags, &c.Position); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (db *DRDatabase) GetCategory(id string) (*Category, error) {
	var c Category
	err := db.pool.QueryRow(context.Background(), `
        SELECT id, display_name, cover_image, tags, position
        FROM categories
        WHERE id = $1
    `, id).Scan(&c.ID, &c.DisplayName, &c.CoverImage, &c.Tags, &c.Position)
	return &c, err
}

func (db *DRDatabase) CreateCategory(c *Category) error {
	_, err := db.pool.Exec(context.Background(), `
        INSERT INTO categories (id, display_name, cover_image, tags, position)
        VALUES ($1, $2, $3, $4, $5)
    `, c.ID, c.DisplayName, c.CoverImage, c.Tags, c.Position)
	return err
}

func (db *DRDatabase) UpdateCategory(id string, c *Category) error {
	_, err := db.pool.Exec(context.Background(), `
        UPDATE categories
        SET display_name = $2, cover_image = $3, tags = $4, position = $5
        WHERE id = $1
    `, id, c.DisplayName, c.CoverImage, c.Tags, c.Position)
	return err
}

func (db *DRDatabase) DeleteCategory(id string) error {
	_, err := db.pool.Exec(context.Background(), `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func (db *DRDatabase) ReorderCategories(orderedIDs []string) error {
	// array_position returns the 1-based index of the id in the array.
	// if not found it returns NULL, which COALESCE turns to 900 to push unlisted categories to the bottom
	_, err := db.pool.Exec(context.Background(), `
        UPDATE categories 
        SET position = COALESCE(array_position($1::text[], id), 900)
    `, orderedIDs)

	return err
}

func (db *DRDatabase) GetImageAuthors() ([][]string, error) {
	rows, err := db.pool.Query(context.Background(), `
        SELECT DISTINCT effective_author, effective_author_url
        FROM images
        WHERE effective_author IS NOT NULL AND effective_author != ''
        ORDER BY effective_author ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors [][]string
	for rows.Next() {
		var author string
		var url *string
		if err := rows.Scan(&author, &url); err != nil {
			return nil, err
		}

		urlStr := ""
		if url != nil {
			urlStr = *url
		}
		authors = append(authors, []string{author, urlStr})
	}

	return authors, nil
}

func (db *DRDatabase) GetCategoryImages(categoryID string, limit, offset int) ([]Image, int, error) {
	var totalCount int
	err := db.pool.QueryRow(context.Background(), `
        SELECT COUNT(*)
        FROM images
        WHERE effective_category_id = $1
    `, categoryID).Scan(&totalCount)

	if err != nil {
		return nil, 0, err
	}

	query := `
        SELECT id, source_id, relative_path, local_path, external_url,
               category_override, author_override, author_url_override, tags_override,
               effective_category_id, effective_author, effective_author_url, effective_tags
        FROM images
        WHERE effective_category_id = $1
        ORDER BY id
        LIMIT $2 OFFSET $3
    `

	rows, err := db.pool.Query(context.Background(), query, categoryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var i Image
		if err := rows.Scan(&i.ID, &i.SourceID, &i.RelativePath, &i.LocalPath, &i.ExternalURL,
			&i.CategoryOverride, &i.AuthorOverride, &i.AuthorURLOverride, &i.TagsOverride,
			&i.EffectiveCategoryID, &i.EffectiveAuthor, &i.EffectiveAuthorURL, &i.EffectiveTags); err != nil {
			return nil, 0, err
		}
		images = append(images, i)
	}

	return images, totalCount, nil
}

// sources & metadata

func (db *DRDatabase) GetSources() ([]Source, error) {
	rows, err := db.pool.Query(context.Background(), `
        SELECT id, name, source_type, root_path, enabled, last_scanned_at
        FROM sources
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var s Source
		if err := rows.Scan(&s.ID, &s.Name, &s.SourceType, &s.RootPath, &s.Enabled, &s.LastScannedAt); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func (db *DRDatabase) GetSource(id int) (*Source, error) {
	var s Source
	err := db.pool.QueryRow(context.Background(), `
        SELECT id, name, source_type, root_path, enabled, last_scanned_at
        FROM sources WHERE id = $1
    `, id).Scan(&s.ID, &s.Name, &s.SourceType, &s.RootPath, &s.Enabled, &s.LastScannedAt)
	return &s, err
}

func (db *DRDatabase) CreateSource(s *Source) error {
	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// insert the source and get its id
	err = tx.QueryRow(context.Background(), `
        INSERT INTO sources (name, source_type, root_path, enabled)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, s.Name, s.SourceType, s.RootPath, s.Enabled).Scan(&s.ID)
	if err != nil {
		return err
	}

	// insert an empty root path_metadata entry for this source
	_, err = tx.Exec(context.Background(), `
        INSERT INTO path_metadata (source_id, relative_path, tags, tag_mode)
        VALUES ($1, '', '{}', 'merge')
    `, s.ID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (db *DRDatabase) UpdateSource(s *Source) error {
	_, err := db.pool.Exec(context.Background(), `
        UPDATE sources
        SET name = $2, source_type = $3, root_path = $4, enabled = $5
        WHERE id = $1
    `, s.ID, s.Name, s.SourceType, s.RootPath, s.Enabled)
	return err
}

func (db *DRDatabase) DeleteSource(id int) error {
	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// reset any category covers that rely on images from this source
	_, err = tx.Exec(context.Background(), `
        UPDATE categories
        SET cover_image = -1
        WHERE cover_image IN (
            SELECT id FROM images WHERE source_id = $1
        )
    `, id)
	if err != nil {
		return err
	}

	// delete the source (which cascades and deletes the images / path_metadata)
	_, err = tx.Exec(context.Background(), `
        DELETE FROM sources WHERE id = $1
    `, id)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (db *DRDatabase) SyncScannedImages(sourceID int, currentPaths []string) error {
	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	if len(currentPaths) > 0 {
		// insert any new images that don't exist yet
		_, err = tx.Exec(context.Background(), `
            INSERT INTO images (source_id, relative_path)
            SELECT $1, t
            FROM unnest($2::text[]) AS t
            ON CONFLICT (source_id, relative_path) DO NOTHING
        `, sourceID, currentPaths)
		if err != nil {
			return err
		}

		// delete any images that no longer exist on disk
		_, err = tx.Exec(context.Background(), `
            DELETE FROM images
            WHERE source_id = $1 AND relative_path != ALL($2::text[])
        `, sourceID, currentPaths)
		if err != nil {
			return err
		}
	} else {
		// if no valid images were found, wipe everything for this source
		_, err = tx.Exec(context.Background(), `
            DELETE FROM images WHERE source_id = $1
        `, sourceID)
		if err != nil {
			return err
		}
	}

	// update the last scanned timestamp
	_, err = tx.Exec(context.Background(), `
        UPDATE sources SET last_scanned_at = NOW() WHERE id = $1
    `, sourceID)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

// returns all metadata path overrides for a given source
func (db *DRDatabase) GetPathMetadataBySource(sourceID int) ([]PathMetadata, error) {
	rows, err := db.pool.Query(context.Background(), `
        SELECT id, source_id, relative_path, category_id, author, author_url, tags, tag_mode
        FROM path_metadata
        WHERE source_id = $1
        ORDER BY relative_path ASC
    `, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []PathMetadata
	for rows.Next() {
		var fm PathMetadata
		if err := rows.Scan(&fm.ID, &fm.SourceID, &fm.RelativePath, &fm.CategoryID, &fm.Author, &fm.AuthorURL, &fm.Tags, &fm.TagMode); err != nil {
			return nil, err
		}
		entries = append(entries, fm)
	}
	return entries, nil
}

// removes a metadata override and recalculates the effective metadata
// for all images, so they fall back to inheriting from their parent folders.
// prevents deleting the root (”) path metadata since it's required for cascades
func (db *DRDatabase) DeletePathMetadata(id int) error {
	var sourceID int
	var relativePath string

	err := db.pool.QueryRow(context.Background(), `
        SELECT source_id, relative_path
        FROM path_metadata
        WHERE id = $1
    `, id).Scan(&sourceID, &relativePath)

	if err != nil {
		return err
	}

	if relativePath == "" {
		return fmt.Errorf("cannot delete the root path metadata for a source")
	}

	_, err = db.pool.Exec(context.Background(), `
        DELETE FROM path_metadata
        WHERE id = $1
    `, id)

	if err != nil {
		return err
	}

	return db.RecalculateEffectiveMetadata(sourceID)
}

// updates directory tags and recalculates all image metadata in that source
func (db *DRDatabase) UpsertPathMetadata(fm *PathMetadata) error {
	_, err := db.pool.Exec(context.Background(), `
        INSERT INTO path_metadata (source_id, relative_path, category_id, author, author_url, tags, tag_mode)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (source_id, relative_path)
        DO UPDATE SET
            category_id = EXCLUDED.category_id,
            author = EXCLUDED.author,
            author_url = EXCLUDED.author_url,
            tags = EXCLUDED.tags,
            tag_mode = EXCLUDED.tag_mode
    `, fm.SourceID, fm.RelativePath, fm.CategoryID, fm.Author, fm.AuthorURL, string(fm.Tags), fm.TagMode)

	if err != nil {
		return err
	}
	return db.RecalculateEffectiveMetadata(fm.SourceID)
}

// for each image, finds the closest path_metadata tree and overrides,
// then update the effective metadata as appropriate
func (db *DRDatabase) RecalculateEffectiveMetadata(sourceID int) error {
	query := `
        WITH closest_paths AS (
            SELECT 
                i.id AS image_id,
                (
                    SELECT fm.id
                    FROM path_metadata fm
                    WHERE fm.source_id = i.source_id
                      AND (fm.relative_path = '' OR i.relative_path LIKE (RTRIM(fm.relative_path, '/') || '/%') OR i.relative_path = fm.relative_path)
                    ORDER BY LENGTH(fm.relative_path) DESC
                    LIMIT 1
                ) AS fm_id
            FROM images i
            WHERE i.source_id = $1
        )
        UPDATE images i
        SET
            effective_category_id = COALESCE(i.category_override, fm.category_id),
            effective_author = COALESCE(i.author_override, fm.author),
            effective_author_url = COALESCE(i.author_url_override, fm.author_url),
            effective_tags = COALESCE(i.tags_override, fm.tags)
        FROM closest_paths closest
        JOIN path_metadata fm ON closest.fm_id = fm.id
        WHERE i.id = closest.image_id
    `
	_, err := db.pool.Exec(context.Background(), query, sourceID)
	return err
}

// images

func (db *DRDatabase) GetImage(id int) (*Image, error) {
	var i Image
	err := db.pool.QueryRow(context.Background(), `
        SELECT id, source_id, relative_path, local_path, external_url,
               category_override, author_override, author_url_override, tags_override,
               effective_category_id, effective_author, effective_author_url, effective_tags
        FROM images WHERE id = $1
    `, id).Scan(&i.ID, &i.SourceID, &i.RelativePath, &i.LocalPath, &i.ExternalURL,
		&i.CategoryOverride, &i.AuthorOverride, &i.AuthorURLOverride, &i.TagsOverride,
		&i.EffectiveCategoryID, &i.EffectiveAuthor, &i.EffectiveAuthorURL, &i.EffectiveTags)
	return &i, err
}

func (db *DRDatabase) UpdateImage(i *Image) error {
	_, err := db.pool.Exec(context.Background(), `
        UPDATE images
        SET category_override = $2, author_override = $3, author_url_override = $4, tags_override = $5, updated_at = NOW()
        WHERE id = $1
    `, i.ID, i.CategoryOverride, i.AuthorOverride, i.AuthorURLOverride, string(i.TagsOverride))

	if err != nil {
		return err
	}
	return db.RecalculateEffectiveMetadata(i.SourceID)
}

func (db *DRDatabase) DeleteImage(id int) error {
	_, err := db.pool.Exec(context.Background(), `DELETE FROM images WHERE id = $1`, id)
	return err
}

// drawing sessions

func (db *DRDatabase) GetSessionImages(categoryID string, tags json.RawMessage, limit int) ([]Image, error) {
	query := `
        SELECT id, source_id, relative_path, local_path, external_url,
               effective_category_id, effective_author, effective_author_url, effective_tags
        FROM images
        WHERE effective_category_id = $1
    `
	args := []interface{}{categoryID}

	if len(tags) > 0 && string(tags) != "[]" && string(tags) != "{}" && string(tags) != "null" {
		query += ` AND effective_tags @> $2::jsonb`
		args = append(args, string(tags))
	}

	query += ` ORDER BY RANDOM() LIMIT $` + strconv.Itoa(len(args)+1)
	args = append(args, limit)

	rows, err := db.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var i Image
		if err := rows.Scan(&i.ID, &i.SourceID, &i.RelativePath, &i.LocalPath, &i.ExternalURL,
			&i.EffectiveCategoryID, &i.EffectiveAuthor, &i.EffectiveAuthorURL, &i.EffectiveTags); err != nil {
			return nil, err
		}
		images = append(images, i)
	}
	return images, nil
}

func (db *DRDatabase) GetSessionImageCount(categoryID string, tags json.RawMessage) (int, error) {
	query := `SELECT COUNT(*) FROM images WHERE effective_category_id = $1`
	args := []interface{}{categoryID}

	if len(tags) > 0 && string(tags) != "[]" && string(tags) != "{}" && string(tags) != "null" {
		query += ` AND effective_tags @> $2::jsonb`
		args = append(args, string(tags))
	}

	var count int
	err := db.pool.QueryRow(context.Background(), query, args...).Scan(&count)
	return count, err
}

func (db *DRDatabase) GetImagesBySourcePath(sourceID int, pathPrefix string, limit int) ([]Image, error) {
	query := `
        SELECT id, source_id, relative_path, local_path, external_url,
               category_override, author_override, author_url_override, tags_override, effective_category_id,
               effective_author, effective_author_url, effective_tags
        FROM images
        WHERE source_id = $1 AND relative_path LIKE $2
        LIMIT $3
    `
	rows, err := db.pool.Query(context.Background(), query, sourceID, pathPrefix+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var i Image
		if err := rows.Scan(&i.ID, &i.SourceID, &i.RelativePath, &i.LocalPath, &i.ExternalURL,
			&i.CategoryOverride, &i.AuthorOverride, &i.AuthorURLOverride, &i.TagsOverride, &i.EffectiveCategoryID,
			&i.EffectiveAuthor, &i.EffectiveAuthorURL, &i.EffectiveTags); err != nil {
			return nil, err
		}
		images = append(images, i)
	}

	return images, nil
}

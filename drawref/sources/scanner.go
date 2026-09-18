package sources

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

type ScannerDB interface {
	SyncScannedImages(sourceID int, currentPaths []string) error
	RecalculateEffectiveMetadata(sourceID int) error
}

var validExtensions = map[string]bool{
	".apng": true,
	".png":  true,
	".avif": true,
	".gif":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
}

func ScanLocalSource(ctx context.Context, sourceID int, rootPath string, db ScannerDB) error {
	var currentPaths []string

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		// Respect request context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// if the root directory is unreadable/missing, fail the entire scan so we don't wipe the database
			if path == rootPath {
				return err
			}
			// skip unreadable sub-folders/files
			return nil
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if validExtensions[ext] {
			rel, err := filepath.Rel(rootPath, path)
			if err == nil {
				// normalize Windows backslashes to forward slashes for the db
				rel = filepath.ToSlash(rel)
				currentPaths = append(currentPaths, rel)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err := db.SyncScannedImages(sourceID, currentPaths); err != nil {
		return err
	}

	if err := db.RecalculateEffectiveMetadata(sourceID); err != nil {
		return err
	}

	return nil
}

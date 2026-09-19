package sources

import (
	"context"
	"io/fs"
	"os"
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

func GetDirectories(ctx context.Context, sourceType string, rootPath string, pathPrefix string) ([]string, error) {
	if sourceType != "local" {
		return []string{}, nil // only local supported for now
	}

	// normalize pathPrefix to local OS format
	localPrefix := filepath.FromSlash(pathPrefix)

	// determine actual directory to read and prefix to match
	var searchDir string
	var matchPrefix string

	if localPrefix == "" || strings.HasSuffix(pathPrefix, "/") {
		searchDir = filepath.Join(rootPath, localPrefix)
		matchPrefix = ""
	} else {
		searchDir = filepath.Join(rootPath, filepath.Dir(localPrefix))
		matchPrefix = filepath.Base(localPrefix)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		// if directory doesn't exist or isn't accessible, just return empty
		return []string{}, nil
	}

	var suggestions []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(matchPrefix)) {
			// construct the relative path to return
			relPath := ""
			if localPrefix == "" || strings.HasSuffix(pathPrefix, "/") {
				relPath = filepath.Join(localPrefix, entry.Name())
			} else {
				relPath = filepath.Join(filepath.Dir(localPrefix), entry.Name())
			}

			// add a trailing slash to indicate it's a directory
			relPath = filepath.ToSlash(relPath) + "/"
			suggestions = append(suggestions, relPath)
		}
	}

	return suggestions, nil
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

func ScanFSSource(ctx context.Context, sourceID int, fsys fs.FS, rootDir string, db ScannerDB) error {
	var currentPaths []string

	err := fs.WalkDir(fsys, rootDir, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil || d.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if validExtensions[ext] {
			// strip the rootDir prefix (e.g., "sample-images/") to get the relative path for the DB
			rel := strings.TrimPrefix(path, rootDir+"/")
			rel = filepath.ToSlash(rel) // fs paths are already forward slashes, but good for safety
			currentPaths = append(currentPaths, rel)
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

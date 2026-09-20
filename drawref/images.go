package drawref

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
)

// uri params

type ImageRequest struct {
	ID int `uri:"id" binding:"required"`
}

// request bodies

type UpdateImageParams struct {
	CategoryOverride  *string         `json:"category_override"`
	AuthorOverride    *string         `json:"author_override"`
	AuthorURLOverride *string         `json:"author_url_override"`
	TagsOverride      json.RawMessage `json:"tags_override"`
}

// only cache thumbnails with these max sizes in pixels
var AllowedThumbnailCacheSizes = []int{300, 600}

func IsAllowedThumbnailCacheSize(size int) bool {
	for _, s := range AllowedThumbnailCacheSizes {
		if s == size {
			return true
		}
	}
	return false
}

// handlers

func serveOriginalImage(c *gin.Context, source *Source, img *Image) {
	if source.SourceType == "local" {
		fullPath := filepath.Join(source.RootPath, img.RelativePath)
		c.File(fullPath)
		return
	}
	if source.SourceType == "samples" {
		fullPath := path.Join("sample-images", img.RelativePath)
		c.FileFromFS(fullPath, http.FS(SamplesFS))
		return
	}
}

func getOriginalImageSize(source *Source, img *Image) (int64, error) {
	if source.SourceType == "local" {
		fullPath := filepath.Join(source.RootPath, img.RelativePath)
		stat, err := os.Stat(fullPath)
		if err != nil {
			return 0, err
		}
		return stat.Size(), nil
	}
	if source.SourceType == "samples" {
		fullPath := path.Join("sample-images", img.RelativePath)
		f, err := SamplesFS.Open(fullPath)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			return 0, err
		}
		return stat.Size(), nil
	}
	return 0, errors.New("unsupported source type")
}

func openOriginalImage(source *Source, img *Image) (io.ReadCloser, error) {
	if source.SourceType == "local" {
		fullPath := filepath.Join(source.RootPath, img.RelativePath)
		return os.Open(fullPath)
	}
	if source.SourceType == "samples" {
		fullPath := path.Join("sample-images", img.RelativePath)
		return SamplesFS.Open(fullPath)
	}
	return nil, errors.New("unsupported source type")
}

func serveImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	img, err := TheDb.GetImage(req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	source, err := TheDb.GetSource(img.SourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch image source"})
		return
	}

	if source.SourceType != "local" && source.SourceType != "samples" {
		if source.SourceType == "s3" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "S3 serving not implemented or URL is missing"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown source type"})
		return
	}

	maxParamStr := c.Query("max")
	maxVal, _ := strconv.Atoi(maxParamStr)

	// if no max, or <= 0, serve full-size original image
	if maxVal <= 0 {
		serveOriginalImage(c, source, img)
		return
	}

	// check full-size file size
	fileSize, err := getOriginalImageSize(source, img)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image file not found"})
		return
	}

	thresholdKB := GetThumbnailMinSizeKB()
	thresholdBytes := int64(thresholdKB) * 1024
	if thresholdBytes > 0 && fileSize <= thresholdBytes {
		// no need to serve a thumbnail
		serveOriginalImage(c, source, img)
		return
	}

	// check if cached thumbnail exists
	canCache := AppConfig.CachePath != "" && IsAllowedThumbnailCacheSize(maxVal)
	var cacheFilePath string
	if canCache {
		cacheFilePath = filepath.Join(AppConfig.CachePath, fmt.Sprintf("%d_%d.jpg", img.ID, maxVal))
		if _, err := os.Stat(cacheFilePath); err == nil {
			c.File(cacheFilePath)
			return
		}
	}

	// open original image stream to thumbnail it
	file, err := openOriginalImage(source, img)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Could not open image file"})
		return
	}

	// decode and resize
	srcImg, err := imaging.Decode(file, imaging.AutoOrientation(true))
	file.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decode image: " + err.Error()})
		return
	}

	// skip resizing and encoding if the requested max size is greater than or equal to
	// the max dimension of the original source image
	srcBounds := srcImg.Bounds()
	if maxVal >= max(srcBounds.Dx(), srcBounds.Dy()) {
		serveOriginalImage(c, source, img)
		return
	}

	// fit scales the image down so that max(width, height) <= maxVal while preserving aspect ratio
	thumb := imaging.Fit(srcImg, maxVal, maxVal, imaging.Lanczos)

	// if image has transparency, composite over white background for clean JPEG output
	bg := imaging.New(thumb.Bounds().Dx(), thumb.Bounds().Dy(), color.White)
	flattened := imaging.Overlay(bg, thumb, image.Pt(0, 0), 1.0)

	// save to cache if enabled & size allowed
	if canCache {
		if err := os.MkdirAll(AppConfig.CachePath, 0755); err == nil {
			tmpFile, err := os.CreateTemp(AppConfig.CachePath, "thumb_*.tmp")
			if err == nil {
				tmpName := tmpFile.Name()
				encodeErr := jpeg.Encode(tmpFile, flattened, &jpeg.Options{Quality: 85})
				tmpFile.Close()
				if encodeErr == nil && os.Rename(tmpName, cacheFilePath) == nil {
					c.File(cacheFilePath)
					return
				}
				_ = os.Remove(tmpName)
			}
		}
	}

	// directly stream if not cached
	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=86400")
	_ = jpeg.Encode(c.Writer, flattened, &jpeg.Options{Quality: 85})
}

func getImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	img, err := TheDb.GetImage(req.ID)
	if err != nil {
		fmt.Println("Could not get image:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, img)
}

func updateImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	var params UpdateImageParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// load base image to ensure existence and keep SourceID etc intact
	img, err := TheDb.GetImage(req.ID)
	if err != nil {
		fmt.Println("Could not find image to update:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// apply any given overrides
	img.CategoryOverride = params.CategoryOverride
	img.AuthorOverride = params.AuthorOverride
	img.AuthorURLOverride = params.AuthorURLOverride
	img.TagsOverride = params.TagsOverride

	// database update triggers metadata recalculation automatically
	err = TheDb.UpdateImage(img)
	if err != nil {
		fmt.Println("Could not update image:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't update image", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "update_image", fmt.Sprintf("Updated image overrides for image %d", req.ID), map[string]interface{}{
		"id": req.ID,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func deleteImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid Image ID must be provided"})
		return
	}

	err := TheDb.DeleteImage(req.ID)
	if err != nil {
		fmt.Println("Could not delete image:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't delete image", "details": err.Error()})
		return
	}

	TheDb.AddLogLine(LogLevelInfo, "delete_image", fmt.Sprintf("Deleted image %d", req.ID), map[string]interface{}{
		"id": req.ID,
	})

	DeleteCachedThumbnails(req.ID)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func getImageAuthors(c *gin.Context) {
	authors, err := TheDb.GetImageAuthors()
	if err != nil {
		fmt.Println("Could not fetch image authors:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch authors"})
		return
	}

	if authors == nil {
		authors = [][]string{}
	}

	c.JSON(http.StatusOK, authors)
}

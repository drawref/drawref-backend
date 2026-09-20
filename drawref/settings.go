package drawref

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

const SettingThumbnailMinFilesizeKB = "thumbnail_min_filesize_kb"
const DefaultThumbnailMinFilesizeKB = 400

var (
	settingsLock             sync.RWMutex
	cachedThumbnailMinSizeKB int = DefaultThumbnailMinFilesizeKB
)

func InitSettings() {
	if TheDb == nil {
		return
	}

	val, err := TheDb.GetSetting(SettingThumbnailMinFilesizeKB, strconv.Itoa(DefaultThumbnailMinFilesizeKB))
	if err == nil {
		if kb, err := strconv.Atoi(val); err == nil && kb >= 0 {
			SetCachedThumbnailMinSizeKB(kb)
			return
		}
	}
	SetCachedThumbnailMinSizeKB(DefaultThumbnailMinFilesizeKB)
}

func GetThumbnailMinSizeKB() int {
	settingsLock.RLock()
	defer settingsLock.RUnlock()
	return cachedThumbnailMinSizeKB
}

func SetCachedThumbnailMinSizeKB(kb int) {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	cachedThumbnailMinSizeKB = kb
}

type SettingsResponse struct {
	ThumbnailMinFilesizeKB int    `json:"thumbnail_min_filesize_kb"`
	CacheEnabled           bool   `json:"cache_enabled"`
	CachePath              string `json:"cache_path,omitempty"`
	AllowedCacheSizes      []int  `json:"allowed_cache_sizes"`
}

type UpdateSettingsRequest struct {
	ThumbnailMinFilesizeKB *int `json:"thumbnail_min_filesize_kb"`
}

func getSettings(c *gin.Context) {
	kb := GetThumbnailMinSizeKB()
	cacheEnabled := AppConfig.CachePath != ""

	c.JSON(http.StatusOK, SettingsResponse{
		ThumbnailMinFilesizeKB: kb,
		CacheEnabled:           cacheEnabled,
		CachePath:              AppConfig.CachePath,
		AllowedCacheSizes:      AllowedThumbnailCacheSizes,
	})
}

func updateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings payload: " + err.Error()})
		return
	}

	if req.ThumbnailMinFilesizeKB != nil {
		if *req.ThumbnailMinFilesizeKB < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "thumbnail_min_filesize_kb must be 0 or greater"})
			return
		}
		kb := *req.ThumbnailMinFilesizeKB
		if err := TheDb.SetSetting(SettingThumbnailMinFilesizeKB, strconv.Itoa(kb)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting: " + err.Error()})
			return
		}
		SetCachedThumbnailMinSizeKB(kb)
	}

	cacheEnabled := AppConfig.CachePath != ""
	c.JSON(http.StatusOK, SettingsResponse{
		ThumbnailMinFilesizeKB: GetThumbnailMinSizeKB(),
		CacheEnabled:           cacheEnabled,
		CachePath:              AppConfig.CachePath,
		AllowedCacheSizes:      AllowedThumbnailCacheSizes,
	})
}

func DeleteCachedThumbnails(imageID int) {
	if AppConfig.CachePath == "" {
		return
	}
	for _, size := range AllowedThumbnailCacheSizes {
		cacheFilePath := fmt.Sprintf("%s/%d_%d.jpg", AppConfig.CachePath, imageID, size)
		_ = removeFileIfExists(cacheFilePath)
	}
}

func removeFileIfExists(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

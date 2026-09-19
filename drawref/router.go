package drawref

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type PingResponse struct {
	OK bool `json:"ok"`
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, PingResponse{OK: true})
}

func GetRouter(trustedProxies []string, corsAllowedFrom []string) (router *gin.Engine) {
	router = gin.Default()

	router.SetTrustedProxies(trustedProxies)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = corsAllowedFrom
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

	// API
	router.GET("/api/ping", ping)
	router.POST("/api/auth", login)

	// serve local images
	router.GET("/image/:id", serveImage)

	// logs
	router.GET("/api/logs", AdminAuthMiddleware(), getLogs)

	// system
	router.GET("/api/system/directories", AdminAuthMiddleware(), autocompleteDirectory)

	// samples
	router.POST("/api/system/load-samples", AdminAuthMiddleware(), loadSamples)

	// categories
	router.GET("/api/categories", getCategories)
	router.POST("/api/categories/reorder", AdminAuthMiddleware(), reorderCategories)
	router.POST("/api/category", AdminAuthMiddleware(), createCategory)
	router.GET("/api/category/:slug", getCategory)
	router.PUT("/api/category/:slug", AdminAuthMiddleware(), editCategory)
	router.DELETE("/api/category/:slug", AdminAuthMiddleware(), deleteCategory)
	router.GET("/api/category/:slug/images", getCategoryImages)

	// sources
	router.GET("/api/sources", AdminAuthMiddleware(), getSources)
	router.POST("/api/source", AdminAuthMiddleware(), createSource)
	router.GET("/api/source/:slug", AdminAuthMiddleware(), getSource)
	router.PUT("/api/source/:slug", AdminAuthMiddleware(), editSource)
	router.DELETE("/api/source/:slug", AdminAuthMiddleware(), deleteSource)
	router.POST("/api/source/:slug/scan", AdminAuthMiddleware(), scanSource)
	router.GET("/api/source/:slug/directories", AdminAuthMiddleware(), getSourceDirectories)

	// source metadata
	router.GET("/api/source/:slug/path-metadata", AdminAuthMiddleware(), getSourcePathMetadata)
	router.POST("/api/source/:slug/path-metadata", AdminAuthMiddleware(), upsertPathMetadata)
	router.PUT("/api/source/:slug/path-metadata", AdminAuthMiddleware(), upsertPathMetadata)
	router.DELETE("/api/source/:slug/path-metadata", AdminAuthMiddleware(), deletePathMetadata)

	// images
	router.GET("/api/authors", getImageAuthors)
	router.GET("/api/image/:id", getImage)
	router.PUT("/api/image/:id", AdminAuthMiddleware(), updateImage)
	router.DELETE("/api/image/:id", AdminAuthMiddleware(), deleteImage)

	// drawing sessions
	router.GET("/api/session", getSession)
	router.GET("/api/session/count", getSessionCount)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found."})
	})
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found."})
	})

	return router
}

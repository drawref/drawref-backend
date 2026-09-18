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

	// logs
	router.GET("/api/logs", AdminAuthMiddleware(), getLogs)

	// categories
	router.GET("/api/categories", getCategories)
	router.POST("/api/category", AdminAuthMiddleware(), createCategory)
	router.GET("/api/category/:slug", getCategory)
	router.PUT("/api/category/:slug", AdminAuthMiddleware(), editCategory)
	router.DELETE("/api/category/:slug", AdminAuthMiddleware(), deleteCategory)

	// sources
	router.GET("/api/sources", AdminAuthMiddleware(), getSources)
	router.POST("/api/source", AdminAuthMiddleware(), createSource)
	router.GET("/api/source/:slug", AdminAuthMiddleware(), getSource)
	router.PUT("/api/source/:slug", AdminAuthMiddleware(), editSource)
	router.DELETE("/api/source/:slug", AdminAuthMiddleware(), deleteSource)
	router.POST("/api/source/:slug/scan", AdminAuthMiddleware(), scanSource)

	// images
	router.GET("/api/image/:slug", getImage)
	router.PUT("/api/image/:slug", AdminAuthMiddleware(), updateImage)
	router.DELETE("/api/image/:slug", AdminAuthMiddleware(), deleteImage)

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

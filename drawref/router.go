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
	// router.POST("/api/auth", login)

	// // DB
	// router.GET("/api/dbs", getYpsDbs)
	// router.GET("/api/db", getLatestYpsDb)
	// router.PUT("/api/db", AdminAuthMiddleware(), updateYpsDb)
	// router.DELETE("/api/db/:slug", AdminAuthMiddleware(), deleteYpsDb)

	// // pages
	// router.GET("/api/page/:slug", getPage)
	// router.PUT("/api/page/:slug", AdminAuthMiddleware(), editPage)

	// // entries
	// router.GET("/api/entry/:slug", getEntry)
	// router.POST("/api/entry/:slug/file", AdminAuthMiddleware(), uploadEntryFile)
	// router.DELETE("/api/entry/:slug/file", AdminAuthMiddleware(), deleteEntryFile)
	// router.GET("/api/browseby", getBrowseByFields)
	// router.GET("/api/search", searchEntries)
	// router.PUT("/api/import-files", AdminAuthMiddleware(), importFileList)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found."})
	})
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found."})
	})

	return router
}

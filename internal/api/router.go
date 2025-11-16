package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter sets up the Gin router with routes and middleware
func NewRouter(handlerDependencies *HandlerDependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.POST("/upload", func(c *gin.Context) {
		handlerDependencies.uploadHandler()(c.Writer, c.Request)
	})
	r.GET("/records", func(c *gin.Context) {
		handlerDependencies.getRecordsHandler()(c.Writer, c.Request)
	})

	return r
}

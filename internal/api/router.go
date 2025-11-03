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

	handler := handlerDependencies.uploadHandler()
	r.POST("/upload", func(c *gin.Context) {
		handler(c.Writer, c.Request)
	})

	return r
}

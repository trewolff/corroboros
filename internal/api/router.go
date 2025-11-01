package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter builds and returns the gin engine. Do not call Run() here.
func NewRouter( /* pass deps e.g. svc timestamp.Service */ ) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// inject dependencies into handler via closure or method receiver
	r.POST("/upload", uploadHandler( /* deps */ ))

	return r
}

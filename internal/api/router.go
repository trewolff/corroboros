package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter sets up the Gin router with routes and middleware
func NewRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// inject dependencies into handler via closure or method receiver
	var maxIntakeSize int64 = 100
	// ...existing code...
	handler := uploadHandler(db, maxIntakeSize) // http.HandlerFunc
	r.POST("/upload", func(c *gin.Context) {
		// adapt http.HandlerFunc to gin.HandlerFunc
		handler(c.Writer, c.Request)
	})
	// ...existing code...

	return r
}

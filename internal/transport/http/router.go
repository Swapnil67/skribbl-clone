package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()

	// * 1. Cors Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// * 2. Load HTML templates & static files (JS/CSS)
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// * 3 UI Views
	r.GET("/", handleHomeView)

	// * 4. Health Checks
	r.GET("/healthz", handleHealthCheck)
	r.GET("/readyz", handleReadinessCheck)

	// * 5. API Endpoints
	v1 := r.Group("/api/v1")
	{
		v1.POST("/session", handleCreateGuestSession)
	}

	return r
}

func handleHomeView(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Game",
	})
}

func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "UP",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func handleReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "READY",
	})
}

func handleCreateGuestSession(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "session endpoint placeholder",
	})
}

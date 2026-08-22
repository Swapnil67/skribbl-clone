package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"skribbl-clone/internal/transport/ws"
)

func NewRouter(manager *ws.RoomManager) *gin.Engine {
	r := gin.New()

	// * 1. Cors Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	wsHandler := ws.NewHandler(manager)

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
		v1.POST("/sessions", handleCreateGuestSession)
	}

	// * WebSocket Endpoint
	r.GET("/ws", wsHandler.HandleWebSocket)

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

// Unable to get relative path between
// file:///Users/apple/Developer/probe/skribble-clone/internal/transport/http/router.go and ; Base path '' must be an absolute path

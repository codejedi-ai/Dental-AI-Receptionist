package api

import (
	"dental-ai-vapi/api/webhook"
	"dental-ai-vapi/config"
	"dental-ai-vapi/db"
	"dental-ai-vapi/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg config.Config, pg *db.Postgres, mongo *db.Mongo) {
	dispatcher := service.NewToolDispatcher(pg, mongo)

	// Vapi webhook endpoint — all tool calls go here
	r.POST("/api/tools", webhook.WebhookHandler(cfg, dispatcher))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "dental-ai-vapi"})
	})

	// GET /api/tools for reachability checks (browser / cellular)
	r.GET("/api/tools", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"ok":      true,
			"service": "dental-ai-vapi-tools",
			"hint":    "Vapi sends POST with JSON body (type tool-calls); include Authorization: Bearer <TOOL_API_KEY>",
		})
	})
}

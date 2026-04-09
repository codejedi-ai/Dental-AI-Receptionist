package api

import (
	"dental-ai-vapi/api/webhook"
	"dental-ai-vapi/config"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg config.Config) {
	// Vapi webhook endpoint
	r.POST("/api/tools", webhook.WebhookHandler(cfg))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "dental-ai-vapi"})
	})

	// GET /api/tools for reachability checks
	r.GET("/api/tools", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"ok":      true,
			"service": "dental-ai-vapi-tools",
			"hint":    "Vapi sends POST with JSON body (type tool-calls); include Authorization: Bearer <TOOL_API_KEY>",
		})
	})
}

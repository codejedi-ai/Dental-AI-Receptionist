package main

import (
	"dental-ai-vapi/api"
	"dental-ai-vapi/config"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using defaults")
	}

	cfg := config.LoadEnvConfig()

	r := gin.Default()
	api.SetupRoutes(r, cfg)

	addr := cfg.Server.ListenAddr
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	log.Printf("Dental-AI-Vapi listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

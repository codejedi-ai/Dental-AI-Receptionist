package main

import (
	"context"
	"log"
	"time"

	"dental-ai-vapi/api"
	"dental-ai-vapi/config"
	"dental-ai-vapi/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using defaults")
	}

	cfg := config.LoadEnvConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize PostgreSQL
	pg, err := db.NewPostgres(ctx, cfg.Database.PostgresURL)
	if err != nil {
		log.Fatalf("❌ PostgreSQL connection failed: %v", err)
	}
	log.Println("✅ PostgreSQL connected")

	// Initialize MongoDB
	mongoURI := cfg.Database.MongoURL
	mongoDB := "dental"
	mongo, err := db.NewMongo(ctx, mongoURI, mongoDB)
	if err != nil {
		log.Fatalf("❌ MongoDB connection failed: %v", err)
	}
	log.Println("✅ MongoDB connected")

	// Seed dentist names if table is empty (development convenience)
	seedDentists(ctx, pg)

	r := gin.Default()
	api.SetupRoutes(r, cfg, pg, mongo)

	addr := cfg.Server.ListenAddr
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	log.Printf("Dental-AI-Vapi listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func seedDentists(ctx context.Context, pg *db.Postgres) {
	// Check if dentists already exist
	names, err := pg.GetAllDentists(ctx)
	if err == nil && len(names) > 0 {
		log.Printf("✅ Dentists loaded from DB: %v", names)
		return
	}

	// Create dentists table if it doesn't exist
	_, err = pg.Pool().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS dentists (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			specialty VARCHAR(100),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("⚠️  Could not create dentists table: %v", err)
		return
	}

	// Insert default dentists
	defaults := []string{"Dr. Michael Park", "Dr. Priya Sharma", "Dr. Sarah Chen"}
	for _, name := range defaults {
		_, _ = pg.Pool().Exec(ctx,
			"INSERT INTO dentists (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", name)
	}
	log.Printf("✅ Dentists seeded: %v", defaults)
}

package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"dental-ai-vapi/internal/db"
	"dental-ai-vapi/internal/handlers"

	"tailscale.com/tsnet"
)

func main() {
	ctx := context.Background()

	pgURL := envOr("DATABASE_URL", "postgres://dental:internal_pg_2024@postgres:5432/dental")
	mongoURL := envOr("MONGO_URL", "mongodb://mongo:27017/dental")
	mongoDB := envOr("MONGO_DB", "dental")
	tailnetName := envOr("TS_HOSTNAME", envOr("TAILSCALE_HOSTNAME", "dental-ai"))
	tailnetAuthKey := envOr("TS_AUTHKEY", envOr("TAILSCALE_AUTHKEY", envOr("TAILSCALE_AUTH_KEY", "")))

	pg, err := db.NewPostgres(ctx, pgURL)
	if err != nil {
		log.Fatalf("PostgreSQL connect: %v", err)
	}
	defer pg.Close()
	log.Println("✅ PostgreSQL connected")

	mongo, err := db.NewMongo(ctx, mongoURL, mongoDB)
	if err != nil {
		log.Fatalf("MongoDB connect: %v", err)
	}
	defer mongo.Close()
	log.Println("✅ MongoDB connected")

	// ─── Tailscale tsnet listener ───
	srv := &tsnet.Server{
		Hostname: tailnetName,
		AuthKey:  tailnetAuthKey,
		Logf:     log.Printf,
	}
	defer srv.Close()

	// Wait until Tailscale is connected
	_, err = srv.Up(ctx)
	if err != nil {
		log.Fatalf("Tailscale connect: %v", err)
	}

	// Get the HTTPS listener with Tailscale's automatic TLS
	ln, err := srv.ListenTLS("tcp", ":443")
	if err != nil {
		// Fallback: plain TCP on port 3000 if TLS fails
		log.Printf("TLS listen failed (%v), falling back to HTTP on :3000", err)
		plainLn, err2 := srv.Listen("tcp", ":3000")
		if err2 != nil {
			log.Fatalf("Tailscale listen: %v", err2)
		}
		serveHTTP(plainLn, pg, mongo, "http://"+tailnetName+":3000")
		return
	}

	// Get hostname for webhook URL
	lc, err := srv.LocalClient()
	if err != nil {
		log.Fatalf("LocalClient: %v", err)
	}
	status, err := lc.Status(ctx)
	if err != nil {
		log.Fatalf("Tailscale status: %v", err)
	}
	self := status.Self
	webhookURL := "https://" + tailnetName + ".ts.net"
	if self.DNSName != "" {
		webhookURL = "https://" + self.DNSName[:len(self.DNSName)-1] // strip trailing dot
	}

	serveHTTP(ln, pg, mongo, webhookURL)
}

func serveHTTP(ln net.Listener, pg *db.Postgres, mongo *db.Mongo, webhookURL string) {
	log.Printf("🦷 Vapi backend listening on %s", ln.Addr())
	log.Printf("   Tailscale webhook: %s/api/tools", webhookURL)

	h := handlers.New(pg, mongo, webhookURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/verify", h.Verify)
	mux.HandleFunc("/api/tools", h.Tools)
	mux.HandleFunc("/webhook", h.Tools)

	if err := http.Serve(ln, mux); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

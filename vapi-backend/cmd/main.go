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

	_, err = srv.Up(ctx)
	if err != nil {
		log.Fatalf("Tailscale connect: %v", err)
	}

	lc, err := srv.LocalClient()
	if err != nil {
		log.Fatalf("LocalClient: %v", err)
	}
	status, err := lc.Status(ctx)
	if err != nil {
		log.Fatalf("Tailscale status: %v", err)
	}
	webhookURL := "https://" + tailnetName + ".ts.net"
	if status.Self.DNSName != "" {
		webhookURL = "https://" + status.Self.DNSName[:len(status.Self.DNSName)-1]
	}

	// Try Funnel first (public internet), fall back to tailnet TLS
	ln, err := srv.ListenFunnel("tcp", ":443")
	if err != nil {
		log.Printf("⚠️  Funnel unavailable (%v) — serving on tailnet TLS only", err)
		ln, err = srv.ListenTLS("tcp", ":443")
		if err != nil {
			log.Fatalf("ListenTLS: %v", err)
		}
	} else {
		log.Println("🌐 Tailscale Funnel active — publicly reachable")
	}

	serveHTTP(ln, pg, mongo, webhookURL)
}

func serveHTTP(ln net.Listener, pg *db.Postgres, mongo *db.Mongo, webhookURL string) {
	log.Printf("🦷 Vapi backend listening on %s", ln.Addr())
	log.Printf("   Webhook URL: %s/api/tools", webhookURL)

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

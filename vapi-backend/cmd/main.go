// Local-only server: run this binary on the same machine as your work session.
// It is not packaged for remote/cloud deployment; expose /api/tools via a tunnel (e.g. Tailscale Funnel) on this host if needed.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"dental-ai-vapi/internal/db"
	"dental-ai-vapi/internal/handlers"
)

func main() {
	ctx := context.Background()

	pgURL := envOr("DATABASE_URL", "postgres://dental:internal_pg_2024@127.0.0.1:5432/dental")
	mongoURL := envOr("MONGO_URL", "mongodb://127.0.0.1:27017/dental")
	mongoDB := envOr("MONGO_DB", "dental")

	if err := enforceLocalDatastore(pgURL, mongoURL); err != nil {
		log.Fatalf("Datastore policy violation: %v", err)
	}

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

	dentists, err := pg.GetAllDentists(ctx)
	if err != nil {
		log.Printf("⚠️  WARNING: Cannot query dentists table: %v", err)
		log.Printf("   check_availability tool will return errors until this is fixed")
	} else if len(dentists) == 0 {
		log.Printf("⚠️  WARNING: Dentists table is empty! Run: docker exec dental-postgres psql -U dental -d dental -f /docker-entrypoint-initdb.d/init.sql")
		log.Printf("   check_availability tool will return errors until seed data is loaded")
	} else {
		log.Printf("✅ Dentists loaded: %v", dentists)
	}

	httpListen := strings.TrimSpace(os.Getenv("HTTP_LISTEN_ADDR"))
	if httpListen == "" {
		httpListen = "127.0.0.1:8080"
		log.Println("HTTP_LISTEN_ADDR unset, using default 127.0.0.1:8080 (loopback only)")
	}
	runHTTP(pg, mongo, httpListen)
}

func runHTTP(pg *db.Postgres, mongo *db.Mongo, addr string) {
	webhookURL := strings.TrimSuffix(strings.TrimSpace(envOr("PUBLIC_BASE_URL", "")), "/")
	if webhookURL == "" {
		webhookURL = localHTTPBaseURL(addr)
	}
	h := handlers.New(pg, mongo, webhookURL)
	mux := newMux(h)
	log.Printf("Vapi backend (local only) listening on %s", addr)
	log.Printf("  Vapi serverUrl: %s/api/tools", webhookURL)
	log.Printf("  Internet access: run a tunnel on this same machine (see tailscale-expose-vapi.sh)")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func localHTTPBaseURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.HasPrefix(addr, ":") {
			return "http://127.0.0.1" + addr
		}
		return "http://" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func newMux(h *handlers.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/verify", h.Verify)
	mux.HandleFunc("/api/tools", h.Tools)
	mux.HandleFunc("/webhook", h.Tools)

	mux.HandleFunc("/api/auth/send-code", h.SendCode)
	mux.HandleFunc("/api/auth/verify-code", h.VerifyCode)
	mux.HandleFunc("/api/auth/session", h.GetSession)

	return mux
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func enforceLocalDatastore(pgURL, mongoURL string) error {
	// Go runs on the host; DBs are expected on loopback (Docker publishes 5432/27017).
	if err := ensureLocalHost(pgURL, map[string]struct{}{
		"localhost":            {},
		"127.0.0.1":            {},
		"host.docker.internal": {},
	}); err != nil {
		return fmt.Errorf("DATABASE_URL must point to local datastore: %w", err)
	}
	if err := ensureLocalHost(mongoURL, map[string]struct{}{
		"localhost":            {},
		"127.0.0.1":            {},
		"host.docker.internal": {},
	}); err != nil {
		return fmt.Errorf("MONGO_URL must point to local datastore: %w", err)
	}
	return nil
}

func ensureLocalHost(raw string, allow map[string]struct{}) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return fmt.Errorf("missing host in URL")
	}
	if _, ok := allow[host]; !ok {
		return fmt.Errorf("host %q is not allowed", host)
	}
	return nil
}

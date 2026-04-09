package config

import (
	"os"
	"strings"
)

type Config struct {
	Server struct {
		ListenAddr string
		PublicURL  string
	}
	Database struct {
		PostgresURL string
		MongoURL    string
	}
	Vapi struct {
		APIKey    string
		PublicKey string
	}
	ToolAPIKey string
	ToolsDebug bool
}

func LoadEnvConfig() Config {
	return Config{
		Server: struct{ ListenAddr, PublicURL string }{
			ListenAddr: getEnv("HTTP_LISTEN_ADDR", "127.0.0.1:8080"),
			PublicURL:  getEnv("PUBLIC_BASE_URL", ""),
		},
		Database: struct{ PostgresURL, MongoURL string }{
			PostgresURL: getEnv("DATABASE_URL", "postgresql://dental:internal_pg_2024@127.0.0.1:5432/dental"),
			MongoURL:    getEnv("MONGO_URL", "mongodb://dental:internal_mongo_2024@127.0.0.1:27017/dental?authSource=dental"),
		},
		Vapi: struct{ APIKey, PublicKey string }{
			APIKey:    getEnv("VAPI_API_KEY", ""),
			PublicKey: getEnv("VAPI_PUBLIC_KEY", ""),
		},
		ToolAPIKey: getEnv("TOOL_API_KEY", ""),
		ToolsDebug: strings.ToLower(strings.TrimSpace(getEnv("VAPI_TOOLS_DEBUG", "0"))) == "1",
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

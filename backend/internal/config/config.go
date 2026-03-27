package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	DiscordClientID    string
	DiscordClientSecret string
	DiscordRedirectURI string
	JWTSecret          string
	FrontendURL        string
	Port               string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tacticarium?sslmode=disable"),
		DiscordClientID:    getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURI: getEnv("DISCORD_REDIRECT_URI", "http://localhost:8080/api/auth/discord/callback"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		Port:               getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

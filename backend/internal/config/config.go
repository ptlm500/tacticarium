package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	JWTSecret           string
	FrontendURL         string
	Port                string

	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string
	AdminGitHubIDs     string
	AdminFrontendURL   string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tacticarium?sslmode=disable"),
		DiscordClientID:     getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURI:  getEnv("DISCORD_REDIRECT_URI", "http://localhost:8080/api/auth/discord/callback"),
		JWTSecret:           getEnv("JWT_SECRET", "dev-secret-change-me"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
		Port:                getEnv("PORT", "8080"),

		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURI:  getEnv("GITHUB_REDIRECT_URI", "http://localhost:8080/api/auth/github/callback"),
		AdminGitHubIDs:     getEnv("ADMIN_GITHUB_IDS", ""),
		AdminFrontendURL:   getEnv("ADMIN_FRONTEND_URL", "http://localhost:5174"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

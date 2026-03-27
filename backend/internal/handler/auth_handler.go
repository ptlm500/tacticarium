package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/config"
)

type AuthHandler struct {
	db      *pgxpool.Pool
	cfg     *config.Config
	discord *auth.DiscordConfig
}

func NewAuthHandler(db *pgxpool.Pool, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:  db,
		cfg: cfg,
		discord: &auth.DiscordConfig{
			ClientID:     cfg.DiscordClientID,
			ClientSecret: cfg.DiscordClientSecret,
			RedirectURI:  cfg.DiscordRedirectURI,
		},
	}
}

func (h *AuthHandler) HandleDiscordRedirect(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.discord.AuthURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleDiscordCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	tokenResp, err := h.discord.ExchangeCode(code)
	if err != nil {
		log.Printf("Discord token exchange error: %v", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Fetch user info
	discordUser, err := auth.FetchDiscordUser(tokenResp.AccessToken)
	if err != nil {
		log.Printf("Discord fetch user error: %v", err)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

	// Upsert user in DB
	displayName := discordUser.GlobalName
	if displayName == "" {
		displayName = discordUser.Username
	}

	var userID string
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO users (discord_id, discord_username, discord_avatar)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (discord_id)
		 DO UPDATE SET discord_username = $2, discord_avatar = $3, updated_at = NOW()
		 RETURNING id`,
		discordUser.ID, displayName, discordUser.Avatar,
	).Scan(&userID)
	if err != nil {
		log.Printf("User upsert error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Generate JWT
	token, err := auth.GenerateToken(h.cfg.JWTSecret, userID, displayName)
	if err != nil {
		log.Printf("JWT generation error: %v", err)
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to frontend
	http.Redirect(w, r, h.cfg.FrontendURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var avatar *string
	var createdAt time.Time
	err := h.db.QueryRow(context.Background(),
		`SELECT discord_avatar, created_at FROM users WHERE id = $1`, user.UserID,
	).Scan(&avatar, &createdAt)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":       user.UserID,
		"username": user.Username,
		"avatar":   avatar,
		"createdAt": createdAt,
	})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

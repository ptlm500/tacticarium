package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/danielgtaylor/huma/v2"
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

// HandleDiscordRedirect stays as a raw chi handler (performs HTTP redirect + sets cookies).
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

// HandleDiscordCallback stays as a raw chi handler (performs HTTP redirect + sets cookies).
func (h *AuthHandler) HandleDiscordCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

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

	tokenResp, err := h.discord.ExchangeCode(code)
	if err != nil {
		slog.ErrorContext(r.Context(), "Discord token exchange error", "error", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	discordUser, err := auth.FetchDiscordUser(tokenResp.AccessToken)
	if err != nil {
		slog.ErrorContext(r.Context(), "Discord fetch user error", "error", err)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

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
		slog.ErrorContext(r.Context(), "User upsert error", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(h.cfg.JWTSecret, userID, displayName)
	if err != nil {
		slog.ErrorContext(r.Context(), "JWT generation error", "error", err)
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", h.cfg.FrontendURL, url.QueryEscape(token))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleMe(ctx context.Context, input *struct{}) (*MeOutput, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	var avatar *string
	var createdAt time.Time
	err := h.db.QueryRow(ctx,
		`SELECT discord_avatar, created_at FROM users WHERE id = $1`, user.UserID,
	).Scan(&avatar, &createdAt)
	if err != nil {
		return nil, huma.Error404NotFound("user not found")
	}

	out := &MeOutput{}
	out.Body.ID = user.UserID
	out.Body.Username = user.Username
	out.Body.Avatar = avatar
	out.Body.CreatedAt = createdAt
	return out, nil
}

// HandleLogout stays as a raw chi handler (needs to set cookies).
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

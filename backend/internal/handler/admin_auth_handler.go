package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/config"
)

type AdminAuthHandler struct {
	cfg           *config.Config
	github        *auth.GitHubConfig
	allowedGitHub map[string]bool
}

func NewAdminAuthHandler(cfg *config.Config) *AdminAuthHandler {
	allowed := make(map[string]bool)
	for _, id := range strings.Split(cfg.AdminGitHubIDs, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = true
		}
	}

	return &AdminAuthHandler{
		cfg: cfg,
		github: &auth.GitHubConfig{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURI:  cfg.GitHubRedirectURI,
		},
		allowedGitHub: allowed,
	}
}

// HandleGitHubRedirect stays as a raw chi handler (performs HTTP redirect + sets cookies).
func (h *AdminAuthHandler) HandleGitHubRedirect(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "github_oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.github.AuthURL(state), http.StatusTemporaryRedirect)
}

// HandleGitHubCallback stays as a raw chi handler (performs HTTP redirect + sets cookies).
func (h *AdminAuthHandler) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("github_oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "github_oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	tokenResp, err := h.github.ExchangeCode(code)
	if err != nil {
		slog.ErrorContext(r.Context(), "GitHub token exchange error", "error", err)
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	ghUser, err := auth.FetchGitHubUser(tokenResp.AccessToken)
	if err != nil {
		slog.ErrorContext(r.Context(), "GitHub fetch user error", "error", err)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

	ghIDStr := strconv.Itoa(ghUser.ID)
	if !h.allowedGitHub[ghIDStr] {
		slog.WarnContext(r.Context(), "Unauthorized admin login attempt", "github_user", ghUser.Login, "github_id", ghIDStr)
		http.Error(w, "forbidden: not an authorized admin", http.StatusForbidden)
		return
	}

	token, err := auth.GenerateTokenWithRole(h.cfg.JWTSecret, ghIDStr, ghUser.Login, "admin")
	if err != nil {
		slog.ErrorContext(r.Context(), "JWT generation error", "error", err)
		http.Error(w, "token generation failed", http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", h.cfg.AdminFrontendURL, url.QueryEscape(token))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *AdminAuthHandler) HandleAdminMe(ctx context.Context, input *struct{}) (*AdminMeOutput, error) {
	admin := auth.GetAdmin(ctx)
	if admin == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	out := &AdminMeOutput{}
	out.Body.GitHubID = admin.GitHubID
	out.Body.GitHubUser = admin.GitHubUser
	return out, nil
}

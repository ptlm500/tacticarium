package auth

import (
	"encoding/json"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// HumaMiddleware returns a huma middleware that validates player JWT tokens.
func HumaMiddleware(jwtSecret string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		tokenString := ""

		authHeader := ctx.Header("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenString == "" {
			tokenString = cookieValue(ctx, "token")
		}

		if tokenString == "" {
			writeError(ctx, 401, "unauthorized")
			return
		}

		claims, err := ValidateToken(jwtSecret, tokenString)
		if err != nil {
			writeError(ctx, 401, "unauthorized")
			return
		}

		ctx = huma.WithValue(ctx, UserContextKey, &UserContext{
			UserID:   claims.UserID,
			Username: claims.Username,
		})
		setUserSpanAttrs(ctx.Context(), claims.UserID, claims.Username)
		next(ctx)
	}
}

// HumaAdminMiddleware returns a huma middleware that validates admin JWT tokens.
func HumaAdminMiddleware(jwtSecret string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		tokenString := ""

		authHeader := ctx.Header("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenString == "" {
			tokenString = cookieValue(ctx, "admin_token")
		}

		if tokenString == "" {
			writeError(ctx, 401, "unauthorized")
			return
		}

		claims, err := ValidateToken(jwtSecret, tokenString)
		if err != nil {
			writeError(ctx, 401, "unauthorized")
			return
		}

		if claims.Role != "admin" {
			writeError(ctx, 403, "forbidden")
			return
		}

		ctx = huma.WithValue(ctx, AdminContextKey, &AdminContext{
			GitHubID:   claims.UserID,
			GitHubUser: claims.Username,
		})
		setUserSpanAttrs(ctx.Context(), claims.UserID, claims.Username)
		next(ctx)
	}
}

// writeError writes a problem details error response without needing the API reference.
func writeError(ctx huma.Context, status int, detail string) {
	ctx.SetStatus(status)
	ctx.SetHeader("Content-Type", "application/problem+json")
	body := map[string]any{
		"status": status,
		"title":  statusText(status),
		"detail": detail,
	}
	_ = json.NewEncoder(ctx.BodyWriter()).Encode(body)
}

func statusText(code int) string {
	switch code {
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	default:
		return "Error"
	}
}

// cookieValue extracts a cookie value from a huma context using the Cookie header.
func cookieValue(ctx huma.Context, name string) string {
	cookie := ctx.Header("Cookie")
	if cookie == "" {
		return ""
	}
	for _, part := range strings.Split(cookie, ";") {
		part = strings.TrimSpace(part)
		if eqIdx := strings.Index(part, "="); eqIdx > 0 {
			if part[:eqIdx] == name {
				return part[eqIdx+1:]
			}
		}
	}
	return ""
}

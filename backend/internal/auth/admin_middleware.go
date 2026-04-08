package auth

import (
	"context"
	"net/http"
	"strings"
)

const AdminContextKey contextKey = "admin"

type AdminContext struct {
	GitHubID   string
	GitHubUser string
}

func AdminMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}

			if tokenString == "" {
				if cookie, err := r.Cookie("admin_token"); err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateToken(jwtSecret, tokenString)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Role != "admin" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), AdminContextKey, &AdminContext{
				GitHubID:   claims.UserID,
				GitHubUser: claims.Username,
			})
			setUserSpanAttrs(ctx, claims.UserID, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAdmin(ctx context.Context) *AdminContext {
	if admin, ok := ctx.Value(AdminContextKey).(*AdminContext); ok {
		return admin
	}
	return nil
}

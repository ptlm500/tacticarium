// Command dev-auth is a local-only authentication helper that mints JWTs
// for fake users so the app can be exercised end-to-end without going through
// Discord OAuth. It is built and run only via docker-compose.dev.yml; nothing
// in the production code path imports or invokes it.
package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/db"
)

const presetsHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Tacticarium dev login</title>
  <style>
    body { font-family: system-ui, sans-serif; background: #0b0d10; color: #d8e1ea;
           display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; }
    .card { background: #14181d; border: 1px solid #2a3038; border-radius: 8px; padding: 32px; width: 360px; }
    h1 { margin: 0 0 8px; font-size: 18px; letter-spacing: 0.15em; text-transform: uppercase; }
    p  { margin: 0 0 24px; font-size: 13px; color: #8a96a3; }
    .row { display: flex; gap: 8px; margin-bottom: 16px; flex-wrap: wrap; }
    a.btn, button { background: #1f2630; color: #d8e1ea; border: 1px solid #2a3038;
                    padding: 8px 14px; border-radius: 4px; text-decoration: none;
                    font-size: 13px; cursor: pointer; }
    a.btn:hover, button:hover { background: #2a3038; }
    form { display: flex; gap: 8px; }
    input { flex: 1; background: #0b0d10; color: #d8e1ea; border: 1px solid #2a3038;
            padding: 8px 10px; border-radius: 4px; font-size: 13px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Dev login</h1>
    <p>Mints a JWT for a fake user and redirects into the app. Local dev only.</p>
    <div class="row">
      {{range .Presets}}<a class="btn" href="/login?username={{.}}">{{.}}</a>{{end}}
    </div>
    <form action="/login" method="get">
      <input name="username" placeholder="custom username" autofocus>
      <button type="submit">Login</button>
    </form>
  </div>
</body>
</html>`

var (
	presets       = []string{"alice", "bob", "charlie"}
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	pageTmpl      = template.Must(template.New("page").Parse(presetsHTML))
)

func main() {
	dbURL := envOr("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/tacticarium?sslmode=disable")
	jwtSecret := envOr("JWT_SECRET", "dev-secret-change-me")
	frontendURL := envOr("FRONTEND_URL", "http://localhost:5173")
	port := envOr("DEV_AUTH_PORT", "8090")

	pool, err := db.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("dev-auth: connect db: %v", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pageTmpl.Execute(w, struct{ Presets []string }{presets})
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			username = "tester"
		}
		if !usernameRegex.MatchString(username) {
			http.Error(w, "username must match [a-zA-Z0-9_-]{1,32}", http.StatusBadRequest)
			return
		}

		discordID := "dev-" + username
		var userID string
		err := pool.QueryRow(r.Context(),
			`INSERT INTO users (discord_id, discord_username)
			 VALUES ($1, $2)
			 ON CONFLICT (discord_id)
			 DO UPDATE SET discord_username = $2, updated_at = NOW()
			 RETURNING id`,
			discordID, username,
		).Scan(&userID)
		if err != nil {
			log.Printf("dev-auth: upsert user %q: %v", username, err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		token, err := auth.GenerateToken(jwtSecret, userID, username)
		if err != nil {
			log.Printf("dev-auth: mint token: %v", err)
			http.Error(w, "token generation failed", http.StatusInternalServerError)
			return
		}

		log.Printf("dev-auth: issued token for %q (user %s)", username, userID)
		http.Redirect(w, r,
			fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, url.QueryEscape(token)),
			http.StatusTemporaryRedirect)
	})

	addr := ":" + port
	log.Printf("dev-auth listening on %s — open http://localhost:%s to log in", addr, port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("dev-auth: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

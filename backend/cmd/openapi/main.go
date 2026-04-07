package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/peter/tacticarium/backend/internal/config"
	"github.com/peter/tacticarium/backend/internal/server"
	"github.com/peter/tacticarium/backend/internal/ws"
)

// main outputs the OpenAPI spec as JSON to stdout.
// It builds the huma API with nil DB pool and a dummy config — no running
// server or database is required. The handlers are never invoked; huma
// only needs the registrations to build the spec.
func main() {
	cfg := &config.Config{
		FrontendURL: "http://localhost:5173",
	}
	hub := ws.NewHub()
	h := server.NewHandlers(nil, hub, cfg)

	r := chi.NewRouter()
	api := server.NewAPI(r, h, "")

	spec, err := json.MarshalIndent(api.OpenAPI(), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling OpenAPI spec: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(spec))
}

# Huma Framework Migration + OTEL Observability Plan



## Context



The backend currently uses raw chi handlers with manual JSON encoding/decoding, plain-text error responses, no OpenAPI spec, and minimal observability (`log.Printf` + chi logger middleware). This migration will:



- Replace all raw chi handlers with huma operations for typed I/O, automatic OpenAPI generation, and RFC 9457 problem details errors

- Convert auth middleware to huma middleware with OpenAPI security scheme annotations

- Add full OTEL observability (HTTP tracing, DB spans, WS/game engine spans, structured logging)

- Enable TypeScript type generation from the OpenAPI spec for both frontends



WebSocket stays as-is (raw chi handler). Existing integration tests must keep passing. Validation stays minimal for now.



---



## Phase 1: Foundation (OTEL + Structured Logging)



Do this first because it's independent of the huma migration and provides immediate value. Subsequent phases benefit from having tracing in place.



### 1.1 OTEL Setup



**New file: `backend/internal/telemetry/telemetry.go`**

- `Init(ctx, serviceName, version) (shutdown func(), err error)`

- Configure OTLP exporter (env-driven: `OTEL_EXPORTER_OTLP_ENDPOINT`, with sensible defaults)

- Set up TracerProvider, MeterProvider

- Register global providers

- Return shutdown function for graceful cleanup



**Modify: `backend/cmd/server/main.go`**

- Call `telemetry.Init()` early in startup

- Defer the shutdown function

- Pass tracer/logger down through dependency chain



### 1.2 Structured Logging



**New file: `backend/internal/logging/logging.go`**

- Configure `slog` with JSON handler

- Add OTEL trace ID extraction middleware/helper so logs correlate with traces

- Export a `FromContext(ctx) *slog.Logger` helper that includes trace info



**Modify: all files using `log.Printf`**

- Replace with `slog.ErrorContext(ctx, ...)` / `slog.InfoContext(ctx, ...)` etc.

- Key files: `game_handler.go`, `ws/hub.go`, `ws/room.go`, `ws/client.go`, `auth_handler.go`



### 1.3 Database Tracing



**Modify: `backend/internal/db/db.go`**

- Add `pgx` OTEL tracing via `github.com/exaring/otelpgx` (or `github.com/XSAM/otelsql` depending on pgx integration)

- Wrap the pgxpool config with the OTEL tracer before creating the pool

- This automatically creates spans for every DB query



### 1.4 HTTP Tracing



**Modify: `backend/internal/server/router.go`**

- Add `otelhttp.NewHandler()` wrapping the chi router (or use `otelhttp` as chi middleware)

- This creates spans for every HTTP request



### 1.5 WebSocket + Game Engine Tracing



**Modify: `backend/internal/ws/room.go`**

- Create spans around action processing (`room.processAction`)

- Include game ID, player number, action type as span attributes



**Modify: `backend/internal/game/engine.go`**

- Create spans around `ApplyAction` calls

- Include action type, game phase as attributes



---



## Phase 2: Huma Setup + Auth Middleware



### 2.1 Huma Adapter Setup



**Modify: `backend/internal/server/router.go`**

- Add `humachi.New(router, humaConfig)` after chi middleware setup

- Configure `huma.DefaultConfig("Tacticarium API", "1.0.0")` with:

  - API description

  - Security schemes (Bearer JWT for player, Bearer JWT for admin)

- The chi router still handles serving; huma wraps it for operation registration

- WebSocket route stays as `router.Get("/ws/game/{gameId}", ...)` on the raw chi router



### 2.2 Auth Middleware Conversion



**Modify: `backend/internal/auth/middleware.go`**

- Add `HumaMiddleware(jwtSecret string) func(ctx huma.Context, next func(huma.Context))`

  - Same JWT validation logic (Bearer header, then cookie fallback)

  - On failure: `huma.WriteErr(api, ctx, 401, "unauthorized")`

  - On success: `ctx = huma.WithValue(ctx, userContextKey, &UserContext{...})` then `next(ctx)`

- Add `HumaAdminMiddleware(jwtSecret string) func(ctx huma.Context, next func(huma.Context))`

  - Same as above + role check

- Keep existing chi middleware functions for now (WebSocket route still needs them via chi group, or handle auth inline as it currently does via query param)

- Add `GetUserFromContext(ctx context.Context) *UserContext` if not already generic enough



### 2.3 Security Scheme Registration



**In router setup:**

```go

api.OpenAPI().Components.SecuritySchemes["bearerAuth"] = &huma.SecurityScheme{

    Type:         "http",

    Scheme:       "bearer",

    BearerFormat: "JWT",

}

```



Operations in protected groups reference this scheme via `huma.Operation{Security: []map[string][]string{{"bearerAuth": {}}}}`



---



## Phase 3: Input/Output Types



### 3.1 Type Organization



**New file: `backend/internal/handler/types.go`**

- All huma input/output structs in one file (they're thin wrappers around existing models)

- Group by domain: auth, factions, missions, games, admin



### 3.2 Input Types Pattern



```go

// Path param inputs

type IDParam struct {

    ID string `path:"id" doc:"Resource ID"`

}



type FactionIDParam struct {

    FactionID string `path:"factionId" doc:"Faction ID"`

}



// Body inputs (for create/update) - reuse existing model types

type CreateFactionInput struct {

    Body models.Faction

}



// Outputs wrap existing models

type FactionOutput struct {

    Body models.Faction

}



type FactionListOutput struct {

    Body []models.Faction

}

```



This approach reuses existing model structs (which already have `json` tags) as the Body type, keeping changes minimal.



### 3.3 Admin CRUD Pattern



The admin handler has 9 resources x 5 operations (45 endpoints). To avoid massive boilerplate, consider a generic registration helper:



```go

func registerCRUD[T any](api huma.API, basePath, tag string, handler *AdminHandler,

    list func(ctx context.Context) ([]T, error),

    get func(ctx context.Context, id string) (*T, error),

    create func(ctx context.Context, item T) (*T, error),

    update func(ctx context.Context, id string, item T) (*T, error),

    del func(ctx context.Context, id string) error,

)

```



However, the current handlers have varied SQL and slightly different patterns. A pragmatic approach:

- Define the input/output types generically where possible

- Register each operation individually but with consistent patterns

- Extract common DB logic into methods that return `(result, error)` instead of writing to `http.ResponseWriter`



### 3.4 Handler Refactoring Pattern



Each handler method needs to change from:

```go

func (h *AdminHandler) ListFactions(w http.ResponseWriter, r *http.Request) {

    rows, err := h.db.Query(r.Context(), sql)

    // ... scan, error handling ...

    writeJSON(w, http.StatusOK, factions)

}

```



To a function registered via `huma.Register`:

```go

func (h *AdminHandler) ListFactions(ctx context.Context, input *struct{}) (*FactionListOutput, error) {

    rows, err := h.db.Query(ctx, sql)

    // ... scan ...

    if err != nil {

        return nil, huma.Error500InternalServerError("database error")

    }

    return &FactionListOutput{Body: factions}, nil

}

```



---



## Phase 4: Migrate Endpoints



Order matters here - do the simplest/most repetitive first to establish the pattern.



### 4.1 Health Check

- Trivial, good first endpoint to verify huma is working

- `huma.Register(api, huma.Operation{...}, func(...) (*HealthOutput, error))`



### 4.2 Player Read-Only Endpoints (Factions + Missions)



**Modify: `backend/internal/handler/faction_handler.go`**

- Convert `ListFactions`, `ListDetachments`, `ListStratagems`, `ListDetachmentStratagems`

- These are simple: path param in, list of models out



**Modify: `backend/internal/handler/mission_handler.go`**

- Convert `ListMissionPacks`, `ListMissions`, `ListSecondaries`, `ListGambits`, `ListMissionRules`, `ListChallengerCards`



### 4.3 Auth Endpoints



**Modify: `backend/internal/handler/auth_handler.go`**

- OAuth redirect/callback handlers are tricky - they do HTTP redirects and set cookies, not JSON responses

- These may need to stay as raw chi handlers OR use `huma.StreamResponse` for redirect control

- `HandleMe` and `HandleLogout` convert cleanly to huma operations



**Modify: `backend/internal/handler/admin_auth_handler.go`**

- Same pattern: OAuth flows stay raw, `HandleAdminMe` converts to huma



### 4.4 Game Endpoints



**Modify: `backend/internal/handler/game_handler.go`**

- Convert `CreateGame`, `ListGames`, `JoinGame`, `GetGame`, `GetGameEvents`, `GetHistory`

- `HandleWebSocket` stays as raw chi handler

- These handlers use auth context - verify `auth.GetUser()` works with huma context



### 4.5 Admin CRUD Endpoints



**Modify: `backend/internal/handler/admin_handler.go`**

- Convert all 45 CRUD endpoints

- Convert 3 bulk import endpoints (multipart file upload - huma supports this via `huma.FormFile` in input structs)

- This is the largest change by line count but most mechanical



### 4.6 Router Rewrite



**Modify: `backend/internal/server/router.go`**

- Replace all `r.Get/Post/Put/Delete` with `huma.Register(api, ...)` calls

- Group operations by tag for clean OpenAPI organization

- Player auth middleware applied via huma middleware to player operations

- Admin auth middleware applied via huma middleware to admin operations

- WebSocket route remains as `router.Get("/ws/game/{gameId}", gameHandler.HandleWebSocket)`

- OAuth redirect routes remain as raw chi handlers



---



## Phase 5: Test Updates



### 5.1 Error Response Format Changes



**Modify: all `*_test.go` files in `backend/internal/handler/`**

- Tests that assert on error response bodies will need updating

- Old: plain text `"not found"`, `"unauthorized"`

- New: JSON problem details `{"status":404,"title":"Not Found","detail":"..."}`

- Update `testutil.DoRequest` or add helpers for asserting problem details responses



### 5.2 Response Structure Verification



- Most success responses should be unchanged (same JSON body)

- Run all tests, fix any that break due to:

  - Error format changes

  - Header changes (Content-Type for errors)

  - Any subtle serialization differences



### 5.3 WebSocket Tests



- Should be unaffected since WS handler stays as raw chi

- Verify they still pass



---



## Phase 6: TypeScript Generation



### 6.1 Tooling



Recommended: **`openapi-typescript`** (npm package)

- Generates TypeScript types from OpenAPI 3.x specs

- Well-maintained, widely used

- Can fetch spec from running server or from file



**Add to both frontends (`frontend/` and `admin/`):**

- Script in `package.json`: `"generate-types": "openapi-typescript http://localhost:8080/openapi.json -o src/types/api.generated.ts"`

- Or generate from saved spec file for CI: export spec to file, commit it, generate types from file



### 6.2 Integration



- Generated types replace manually maintained type files in `frontend/src/types/` and `admin/src/api/admin.ts`

- API client functions updated to use generated types

- This is a follow-up task after the migration is stable



---



## Files Summary



### New Files

| File | Purpose |

|------|---------|

| `backend/internal/telemetry/telemetry.go` | OTEL provider init, shutdown |

| `backend/internal/logging/logging.go` | slog setup, trace-correlated logger |

| `backend/internal/handler/types.go` | All huma input/output structs |



### Modified Files

| File | Changes |

|------|---------|

| `backend/cmd/server/main.go` | OTEL init, structured logging setup |

| `backend/internal/server/router.go` | Huma adapter, rewrite all route registrations |

| `backend/internal/auth/middleware.go` | Add huma middleware variants |

| `backend/internal/handler/admin_handler.go` | Convert all handlers to huma signature |

| `backend/internal/handler/game_handler.go` | Convert REST handlers (keep WS raw) |

| `backend/internal/handler/faction_handler.go` | Convert all handlers |

| `backend/internal/handler/mission_handler.go` | Convert all handlers |

| `backend/internal/handler/auth_handler.go` | Convert me/logout; keep OAuth raw |

| `backend/internal/handler/admin_auth_handler.go` | Convert admin me; keep OAuth raw |

| `backend/internal/handler/helpers.go` | May be removable (huma handles JSON) |

| `backend/internal/db/db.go` | Add pgx OTEL tracing |

| `backend/internal/ws/room.go` | Add action processing spans |

| `backend/internal/game/engine.go` | Add game engine spans |

| `backend/internal/handler/*_test.go` | Update error response assertions |

| `backend/go.mod` | Add huma, otelpgx, otelhttp deps |



### Dependencies to Add

```

github.com/danielgtaylor/huma/v2

github.com/danielgtaylor/huma/v2/adapters/humachi

go.opentelemetry.io/otel (already in go.mod)

go.opentelemetry.io/otel/sdk

go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp

go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp (already in go.mod)

github.com/exaring/otelpgx

go.opentelemetry.io/contrib/bridges/otelslog (slog-to-otel bridge)

```



---



## Verification



1. **Build:** `go build ./...` passes

2. **Tests:** `go test ./...` - all existing integration tests pass (with updated error assertions)

3. **OpenAPI spec:** Start server, `curl http://localhost:8080/openapi.json` returns valid spec

4. **Spec validation:** Paste spec into editor.swagger.io or use `openapi-generator validate`

5. **Manual smoke test:** Hit a few endpoints via curl, verify JSON responses and problem details errors

6. **Tracing:** Start with OTEL collector (or Jaeger), make requests, verify traces appear with HTTP + DB spans

7. **Logging:** Verify structured JSON logs with trace IDs

8. **Frontend compatibility:** Both frontends can still call the API without changes (response shapes unchanged for success cases)

9. **TypeScript generation:** Run `openapi-typescript` against the spec, verify types look correct


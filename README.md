# DashAI Go Gateway

Go shared backend gateway. Multiple Go API modules share a single deployment on Render.

AI-assisted development with Claude Code.

## Architecture

```
dashai-go/
├── cmd/server/main.go              # HTTP entry (chi router)
├── internal/
│   ├── config/config.go            # envconfig
│   ├── middleware/                  # CORS, security, rate limit, request ID
│   │   └── auth/jwt.go             # Logto JWT (shared, per-module opt-in)
│   ├── database/database.go        # pgx connection pool (Neon PostgreSQL)
│   ├── response/response.go        # {success, data, error} envelope
│   └── demo/router.go              # Demo module (/demo)
├── Dockerfile
├── render.yaml
└── go.mod
```

## Adding a New Module

1. Create `internal/yourmodule/router.go`:

```go
package yourmodule

func Router(cfg *config.Config, db *database.DB) chi.Router {
    r := chi.NewRouter()
    r.Get("/api/items", handleItems(db))
    return r
}
```

2. Mount in `cmd/server/main.go`:

```go
r.Mount("/yourmodule", yourmodule.Router(cfg, db))
```

## Run Locally

```bash
# Without database
go run ./cmd/server/

# With database
DATABASE_URL="postgresql://..." go run ./cmd/server/

# Debug mode (adds localhost CORS origins)
DEBUG=true go run ./cmd/server/
```

Port: 8101 (default)

## Endpoints

| Path | Method | Description |
|------|--------|-------------|
| `/health` | GET/HEAD | Health check (UptimeRobot) |
| `/` | GET | Gateway info + service list |
| `/demo/api/ping` | GET | Pong + timestamp |
| `/demo/api/status` | GET | Module + DB status |
| `/demo/api/protected` | GET | Requires Logto JWT |

## Deploy (Render)

Render auto-deploys on `git push` to main. See `render.yaml`.

## Docker

```bash
docker build -t dashai-go .
docker run -p 8101:8101 dashai-go
```

## Tech Stack

| Component | Choice |
|-----------|--------|
| Router | chi v5 |
| Config | envconfig |
| Database | pgx v5 (Neon PostgreSQL) |
| Auth | golang-jwt v5 (Logto JWKS) |
| Rate Limit | httprate |
| Logging | slog |

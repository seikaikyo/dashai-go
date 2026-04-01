---
title: DashAI Go Shared Backend
type: feature
status: completed
created: 2026-04-01
---

# DashAI Go Shared Backend

## Background

dashai-api (Python/FastAPI) 已運行 5 個子模組，部署在 Render Singapore。
現在需要一個 Go 版的共用後端，用相同的 gateway 模式，讓未來 Go API 服務
共用一個 Render 部署，降低營運成本。

Go 專案目前有 go-edge-gateway 和 go-factory-io，但它們是工廠現場的獨立
binary，不是 HTTP API 服務。dashai-go 是給 HTTP API 用的共用後端。

## Architecture

```
dashai-go/
├── cmd/
│   └── server/
│       └── main.go                 # HTTP 入口 (chi router)
├── internal/
│   ├── config/
│   │   └── config.go               # 統一環境變數 (envconfig)
│   ├── middleware/
│   │   ├── cors.go                 # CORS middleware
│   │   ├── security.go             # Security headers
│   │   ├── ratelimit.go            # Rate limiting
│   │   ├── requestid.go            # X-Request-ID
│   │   └── auth/
│   │       └── jwt.go              # Logto JWT 驗證 (共用，各模組選用)
│   ├── database/
│   │   └── database.go             # Neon PostgreSQL 連線池 (pgx)
│   ├── response/
│   │   └── response.go             # 統一回應格式 {success, data, error}
│   └── demo/
│       └── router.go               # 示範模組 /demo
├── render.yaml                     # Render 部署設定
├── go.mod
├── go.sum
├── Dockerfile
├── LICENSE
└── README.md
```

### 對照 dashai-api

| 層級 | dashai-api (Python) | dashai-go (Go) |
|------|--------------------|--------------------|
| 入口 | `main.py` FastAPI | `cmd/server/main.go` chi |
| Config | `config.py` pydantic | `internal/config/config.go` envconfig |
| DB | `database.py` SQLModel | `internal/database/database.go` pgx |
| CORS | FastAPI CORSMiddleware | `internal/middleware/cors.go` chi/cors |
| Auth | 各模組 `middleware/auth.py` | `internal/middleware/auth/jwt.go` 共用 |
| 回應 | `{success, data, error}` | `internal/response/response.go` |
| 子模組 | `factory/router.py` | `internal/demo/router.go` |

## Tech Stack

| 項目 | 選擇 | 理由 |
|------|------|------|
| HTTP Router | `chi` v5 | 輕量、支援 sub-router mounting、Go 生態主流 |
| Config | `kelseyhightower/envconfig` | 簡單、env-only、不需 YAML |
| DB | `jackc/pgx` v5 | Go PostgreSQL 標準、連線池內建 |
| JWT | `golang-jwt/jwt` v5 | 搭配 Logto JWKS 驗證 |
| Rate Limit | `go-chi/httprate` | chi 生態、token bucket |
| Logging | `log/slog` | Go 標準庫、結構化 |

## Core Components

### Config (envconfig)

```go
type Config struct {
    Port        int    `envconfig:"PORT" default:"8101"`
    Debug       bool   `envconfig:"DEBUG" default:"false"`
    DatabaseURL string `envconfig:"DATABASE_URL"`

    // Logto JWT (共用，各模組選用)
    LogtoEndpoint    string `envconfig:"LOGTO_ENDPOINT"`
    LogtoAPIResource string `envconfig:"LOGTO_API_RESOURCE"`

    // CORS
    CORSOrigins []string `envconfig:"CORS_ORIGINS" default:""`

    // Rate Limiting
    RateLimit int `envconfig:"RATE_LIMIT" default:"30"` // per minute
}
```

### 統一回應格式

```go
// 與 dashai-api 一致
type Response struct {
    Success bool   `json:"success"`
    Data    any    `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
    Total   int    `json:"total,omitempty"`
    Page    int    `json:"page,omitempty"`
}

func OK(w http.ResponseWriter, data any)
func OKPage(w http.ResponseWriter, data any, total, page int)
func Err(w http.ResponseWriter, status int, msg string)
```

### 子模組掛載

```go
// cmd/server/main.go
r := chi.NewRouter()

// 共用 middleware
r.Use(middleware.RequestID)
r.Use(middleware.SecurityHeaders)
r.Use(middleware.CORS(cfg))
r.Use(middleware.RateLimit(cfg))

// Health + Root
r.Get("/health", handlers.Health)
r.Get("/", handlers.Root)

// 子模組掛載 (與 dashai-api 的 include_router 對應)
r.Mount("/demo", demo.Router(cfg, db))
// 未來: r.Mount("/xxx", xxx.Router(cfg, db))
```

### 子模組 Pattern

```go
// internal/demo/router.go
func Router(cfg *config.Config, db *pgxpool.Pool) chi.Router {
    r := chi.NewRouter()

    // 模組專屬 middleware (選用)
    // r.Use(auth.RequireJWT(cfg))

    r.Get("/api/ping", handlePing)
    r.Get("/api/items", handleListItems(db))

    return r
}
```

### DB 連線池

```go
// internal/database/database.go
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    // pool_max_conns=8, pool_min_conns=2
    // health check: pool.Ping(ctx)
    return pgxpool.NewWithConfig(ctx, config)
}
```

## Demo Module

第一個示範模組 `/demo`，驗證完整鏈路可用：

| Endpoint | 方法 | 說明 | 認證 |
|----------|------|------|------|
| `/demo/api/ping` | GET | 回 pong + timestamp | 無 |
| `/demo/api/items` | GET | 從 DB 讀 items (如有 DB) | 無 |
| `/demo/api/protected` | GET | 需 JWT token | Logto JWT |

## Render Deployment

```yaml
services:
  - type: web
    name: dashai-go
    runtime: go
    region: singapore
    plan: starter
    branch: main
    buildCommand: go build -o server ./cmd/server/
    startCommand: ./server
    healthCheckPath: /health
    envVars:
      - key: GO_VERSION
        value: 1.26.1
```

## Port

| 專案 | Port |
|------|------|
| dashai-go (本機) | 8101 |
| dashai-api (Python) | 8000-8010 |

加入 ecosystem.config.cjs：
```js
{ name: 'dashai-go', cwd: '~/github/dashai-go', script: 'go', args: 'run ./cmd/server/', env: { PORT: 8101 } }
```

## Impact

| Item | Detail |
|------|--------|
| New repo | `github.com/seikaikyo/dashai-go` |
| New files | ~15 Go files |
| Dependencies | chi, pgx, envconfig, golang-jwt, httprate |
| Deployment | Render Singapore, starter plan |
| Port | 8101 (本機) |

## Implementation Plan

### Phase 1: Skeleton + Middleware
1. `go.mod` + 依賴
2. `internal/config/config.go` — envconfig
3. `internal/middleware/` — CORS, security headers, request ID, rate limit
4. `internal/response/response.go` — 統一回應格式
5. `cmd/server/main.go` — chi router + health + root

### Phase 2: Database
1. `internal/database/database.go` — pgx 連線池
2. 可選，沒有 DB 需求時 graceful skip

### Phase 3: Auth
1. `internal/middleware/auth/jwt.go` — Logto JWKS RS256 驗證
2. JWKS cache (1hr TTL)

### Phase 4: Demo Module
1. `internal/demo/router.go` — 示範完整模組 pattern
2. ping / items / protected endpoints

### Phase 5: Deployment
1. `Dockerfile` (multi-stage build)
2. `render.yaml`
3. `README.md`
4. GitHub Actions CI

## Test Plan

| Phase | Test |
|-------|------|
| Skeleton | `curl localhost:8101/health` 回 `{"status":"ok"}` |
| Middleware | 回應含 security headers + X-Request-ID |
| DB | `curl /demo/api/items` 回空陣列 (DB 連上) |
| Auth | `curl /demo/api/protected` 無 token 回 401 |
| Docker | `docker build` + `docker run` + health check |
| Render | 部署後 `/health` 正常 |

## Checklist

- [x] Phase 1: Skeleton + Middleware
- [x] Phase 2: Database (pgx, graceful skip)
- [x] Phase 3: Auth (Logto JWT)
- [x] Phase 4: Demo Module
- [x] Phase 5: Deployment config
- [x] README
- [x] GitHub Actions CI

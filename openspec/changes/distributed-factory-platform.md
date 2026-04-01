---
title: 分散式工廠平台 (Distributed Factory Platform)
type: feature
status: planned
created: 2026-04-01
---

# 分散式工廠平台

dashai-go 從單純的 REST API gateway 升級為工廠分散式系統的 cloud coordinator，
搭配 go-edge-gateway (edge agent) 和 go-ot-security (security agent) 形成完整的
工廠 IoT + 資安平台。

## 前置作業：Render 部署 (Phase 0)

Go 1.26 Docker image 已上架，三個專案可以部署。

### dashai-go (已有 render.yaml)

| 項目 | 值 |
|------|---|
| Repo | github.com/seikaikyo/dashai-go |
| Render Service | srv-d76ao27fte5s73efs92g (已暫停) |
| Plan | Starter → 功能驗證後降 Hobby |
| Port | 8101 |
| Region | Singapore |
| Build | `go build -ldflags="-s -w" -o server ./cmd/server/` |
| Health | `/health` |
| DB | Neon PostgreSQL (patient-cell-76478954, Singapore) |

**步驟：**
1. `cd ~/github/dashai-go && git push` (有 1 筆未 push commit)
2. Render Dashboard → Resume service → Enable auto deploy
3. 手動 trigger deploy 或等 webhook
4. `curl https://dashai-go.onrender.com/health` 驗證

### go-edge-gateway (新建 Render service)

| 項目 | 值 |
|------|---|
| Repo | github.com/seikaikyo/go-edge-gateway |
| Plan | Starter → 降 Hobby |
| Port | 8080 |
| Region | Singapore |
| Runtime | Docker (Dockerfile 已有) |
| Mode | Demo mode (無實體設備，模擬資料) |

**問題：** Dockerfile 用 `replace` directive 引用 go-factory-io 本地路徑。
部署前需改為 GitHub module path。

**步驟：**
1. go.mod 移除 `replace` directive，改用 GitHub tag/commit
2. go-factory-io 先 push tag (v0.1.0)
3. edge-gateway `go mod tidy` 確認能拉到
4. 新增 render.yaml
5. 加入 demo config (edge-gateway.demo.yaml) — 不連實體設備，走模擬
6. Render Dashboard → New Web Service → 連 GitHub repo
7. 驗證 health endpoint

### go-factory-io (新建 Render service)

| 項目 | 值 |
|------|---|
| Repo | github.com/seikaikyo/go-factory-io (org: dashfactory) |
| Plan | Starter → 降 Hobby |
| Port | 10000 |
| Region | Singapore |
| Runtime | Docker (Dockerfile 已有) |
| Command | `secsgem studio --port 10000` |

**步驟：**
1. Render Dashboard → New Web Service → Docker
2. 環境變數: PORT=10000
3. 驗證 Studio UI 可存取

### Phase 0 完成標準
- [ ] dashai-go `/health` 回 200
- [ ] go-edge-gateway `/health` 回 200
- [ ] go-factory-io Studio UI 可開啟
- [ ] 三個都跑在 Starter plan
- [ ] UptimeRobot 加入監控

---

## Phase 1: Edge 註冊 + 心跳 (Edge Registration)

dashai-go 新增 `/edge` 路由，讓 edge 節點自我註冊和回報狀態。

### API 設計

```
POST   /edge/register     # Edge 節點註冊
POST   /edge/heartbeat    # 心跳回報 (每 30s)
GET    /edge/nodes         # 列出所有節點
GET    /edge/nodes/{id}    # 單一節點詳情
DELETE /edge/nodes/{id}    # 移除節點
```

#### POST /edge/register

Request:
```json
{
  "node_id": "edge-gw-factory-a-01",
  "node_type": "edge-gateway",
  "version": "0.1.0",
  "location": "Factory A, Line 1",
  "capabilities": ["modbus", "secsgem", "mqtt"],
  "endpoints": {
    "health": "http://192.168.1.10:8080/health",
    "api": "http://192.168.1.10:8080/api"
  },
  "metadata": {
    "os": "linux/arm64",
    "uptime_seconds": 3600
  }
}
```

Response:
```json
{
  "success": true,
  "data": {
    "node_id": "edge-gw-factory-a-01",
    "registered_at": "2026-04-01T12:00:00Z",
    "heartbeat_interval_ms": 30000,
    "coordinator_version": "0.2.0"
  }
}
```

#### POST /edge/heartbeat

Request:
```json
{
  "node_id": "edge-gw-factory-a-01",
  "status": "healthy",
  "uptime_seconds": 7200,
  "plugins": {
    "modbus": { "status": "running", "devices": 3 },
    "secsgem": { "status": "running", "devices": 1 },
    "mqtt": { "status": "stopped" }
  },
  "system": {
    "cpu_percent": 12.5,
    "memory_mb": 45,
    "goroutines": 28
  }
}
```

Response:
```json
{
  "success": true,
  "data": {
    "ack": true,
    "commands": []
  }
}
```

#### GET /edge/nodes

Response:
```json
{
  "success": true,
  "data": [
    {
      "node_id": "edge-gw-factory-a-01",
      "node_type": "edge-gateway",
      "status": "online",
      "last_heartbeat": "2026-04-01T12:05:00Z",
      "location": "Factory A, Line 1",
      "capabilities": ["modbus", "secsgem", "mqtt"],
      "device_count": 4,
      "version": "0.1.0"
    }
  ],
  "total": 1
}
```

### DB Schema

```sql
CREATE TABLE edge_nodes (
  id            TEXT PRIMARY KEY,
  node_type     TEXT NOT NULL,       -- edge-gateway | ot-security
  version       TEXT,
  location      TEXT,
  capabilities  JSONB DEFAULT '[]',
  endpoints     JSONB DEFAULT '{}',
  metadata      JSONB DEFAULT '{}',
  status        TEXT DEFAULT 'online',  -- online | offline | degraded
  last_heartbeat TIMESTAMPTZ,
  registered_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_edge_nodes_status ON edge_nodes(status);
CREATE INDEX idx_edge_nodes_type ON edge_nodes(node_type);
```

### dashai-go 新增檔案

```
internal/
  edge/
    router.go        # chi sub-router, mount at /edge
    handler.go       # Register, Heartbeat, ListNodes, GetNode, DeleteNode
    store.go         # PostgreSQL CRUD (edge_nodes table)
    model.go         # Node, RegisterRequest, HeartbeatRequest structs
    monitor.go       # Background goroutine: 90s 沒心跳 → 標記 offline
```

### go-edge-gateway Agent 端改動

```
core/
  agent.go           # 新增: Coordinator client
                     #   - 啟動時 POST /edge/register
                     #   - 每 30s POST /edge/heartbeat
                     #   - 收集 plugin health + system metrics
  config.go          # 新增: coordinator 區塊
                     #   coordinator:
                     #     url: "https://dashai-go.onrender.com"
                     #     node_id: "edge-gw-factory-a-01"
                     #     location: "Factory A, Line 1"
                     #     heartbeat_interval: 30s
```

### Phase 1 完成標準
- [ ] dashai-go `/edge/nodes` 回空陣列
- [ ] edge-gw 啟動後自動註冊
- [ ] edge-gw 每 30s 心跳，dashai-go DB 更新
- [ ] 90s 沒心跳 → 狀態變 offline
- [ ] curl 測試所有 5 個 endpoint

---

## Phase 2: 設備事件匯流 (Event Ingestion)

Edge 節點把設備事件 POST 到 dashai-go，dashai-go 存儲 + WebSocket 即時推送。

### API 設計

```
POST   /events/ingest        # 批量事件上報
GET    /events                # 查詢事件 (?node_id=&type=&limit=&offset=)
GET    /events/stream         # WebSocket 即時事件流
GET    /events/stats          # 事件統計
```

#### POST /events/ingest

Request:
```json
{
  "node_id": "edge-gw-factory-a-01",
  "events": [
    {
      "event_id": "evt-uuid-1",
      "timestamp": "2026-04-01T12:00:01Z",
      "source": "modbus:192.168.1.100",
      "type": "register_change",
      "severity": "info",
      "data": {
        "address": 40001,
        "old_value": 100,
        "new_value": 150,
        "device_name": "siemens-plc-01"
      }
    },
    {
      "event_id": "evt-uuid-2",
      "timestamp": "2026-04-01T12:00:02Z",
      "source": "secsgem:eq-01",
      "type": "state_change",
      "severity": "info",
      "data": {
        "from": "IDLE",
        "to": "PROCESSING",
        "lot_id": "LOT-20260401-001"
      }
    }
  ]
}
```

Response:
```json
{
  "success": true,
  "data": { "accepted": 2, "rejected": 0 }
}
```

#### GET /events/stream (WebSocket)

升級為 WebSocket 後，server push 即時事件：
```json
{
  "event_id": "evt-uuid-3",
  "node_id": "edge-gw-factory-a-01",
  "timestamp": "2026-04-01T12:00:05Z",
  "source": "modbus:192.168.1.100",
  "type": "alarm",
  "severity": "high",
  "data": { "message": "Temperature exceeded threshold" }
}
```

### DB Schema

```sql
CREATE TABLE events (
  id          TEXT PRIMARY KEY,
  node_id     TEXT NOT NULL REFERENCES edge_nodes(id),
  timestamp   TIMESTAMPTZ NOT NULL,
  source      TEXT NOT NULL,
  type        TEXT NOT NULL,
  severity    TEXT DEFAULT 'info',  -- info | warning | high | critical
  data        JSONB DEFAULT '{}',
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_events_node ON events(node_id, timestamp DESC);
CREATE INDEX idx_events_type ON events(type, timestamp DESC);
CREATE INDEX idx_events_severity ON events(severity) WHERE severity != 'info';
```

### dashai-go 新增檔案

```
internal/
  events/
    router.go        # chi sub-router, mount at /events
    handler.go       # Ingest, List, Stream (WebSocket), Stats
    store.go         # PostgreSQL CRUD + batch insert
    model.go         # Event, IngestRequest structs
    websocket.go     # WebSocket hub: broadcast to connected clients
```

### go-edge-gateway Agent 端改動

```
core/
  uplink.go          # 修改: 新增 coordinator uplink type
                     #   除了 MQTT/file/stdout，加 "coordinator" 選項
                     #   batch 收集 plugin 訊息，每 5s POST /events/ingest
  config.go          # uplink 新增 coordinator type
                     #   uplink:
                     #     type: coordinator
                     #     batch_interval: 5s
                     #     max_batch_size: 100
```

### Phase 2 完成標準
- [ ] edge-gw 設備事件 batch POST 到 dashai-go
- [ ] dashai-go `/events` 查詢正常（分頁、篩選）
- [ ] WebSocket `/events/stream` 即時收到事件
- [ ] `/events/stats` 回傳各類型事件計數
- [ ] 壓力測試: 100 events/batch 正常處理

---

## Phase 3: OT 資安匯總 (Security Aggregation)

go-ot-security 以 agent mode 運行，掃描結果 POST 到 dashai-go 匯總。

### API 設計

```
POST   /security/report       # 掃描報告上傳
GET    /security/reports       # 歷史報告列表
GET    /security/reports/{id}  # 報告詳情
GET    /security/dashboard     # 資安儀表板 (匯總所有節點)
GET    /security/alerts        # 跨節點告警
POST   /security/alerts/{id}/ack  # 確認告警
```

#### POST /security/report

Request:
```json
{
  "node_id": "ot-sec-factory-a",
  "scan_id": "scan-uuid-1",
  "timestamp": "2026-04-01T12:00:00Z",
  "subnet": "192.168.1.0/24",
  "summary": {
    "total_devices": 47,
    "ot_devices": 12,
    "it_devices": 35,
    "critical_vulns": 3,
    "high_vulns": 8,
    "it_ot_separated": false
  },
  "compliance": {
    "iec62443": { "passed": 7, "total": 10, "score": 70 },
    "nist_csf": { "passed": 5, "total": 7, "score": 71 },
    "iso27001": { "passed": 4, "total": 7, "score": 57 },
    "semi_e187": { "passed": 4, "total": 5, "score": 80 }
  },
  "alerts": [
    {
      "severity": "critical",
      "type": "new_device",
      "message": "New device detected: 192.168.1.45",
      "technique": "T0842",
      "device_ip": "192.168.1.45"
    }
  ],
  "devices": [
    {
      "ip": "192.168.1.100",
      "vendor": "Siemens",
      "type": "plc",
      "protocols": ["s7comm"],
      "risk_score": 6.5,
      "vulns": ["CVE-2019-13945"]
    }
  ]
}
```

#### GET /security/dashboard

Response:
```json
{
  "success": true,
  "data": {
    "nodes": 3,
    "last_scan": "2026-04-01T12:00:00Z",
    "total_devices": 142,
    "total_ot": 38,
    "total_it": 104,
    "critical_alerts": 5,
    "overall_compliance": {
      "iec62443": 68,
      "nist_csf": 72,
      "iso27001": 61,
      "semi_e187": 75
    },
    "nodes_detail": [
      {
        "node_id": "ot-sec-factory-a",
        "location": "Factory A",
        "devices": 47,
        "critical": 3,
        "last_scan": "2026-04-01T12:00:00Z",
        "status": "online"
      }
    ]
  }
}
```

### DB Schema

```sql
CREATE TABLE security_reports (
  id          TEXT PRIMARY KEY,
  node_id     TEXT NOT NULL REFERENCES edge_nodes(id),
  scan_id     TEXT NOT NULL,
  timestamp   TIMESTAMPTZ NOT NULL,
  subnet      TEXT,
  summary     JSONB NOT NULL,
  compliance  JSONB NOT NULL,
  devices     JSONB DEFAULT '[]',
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE security_alerts (
  id          TEXT PRIMARY KEY,
  report_id   TEXT REFERENCES security_reports(id),
  node_id     TEXT NOT NULL,
  severity    TEXT NOT NULL,
  type        TEXT NOT NULL,
  message     TEXT NOT NULL,
  technique   TEXT,
  device_ip   TEXT,
  ack         BOOLEAN DEFAULT FALSE,
  ack_at      TIMESTAMPTZ,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_security_reports_node ON security_reports(node_id, timestamp DESC);
CREATE INDEX idx_security_alerts_unack ON security_alerts(ack, severity) WHERE ack = FALSE;
```

### dashai-go 新增檔案

```
internal/
  security/
    router.go        # chi sub-router, mount at /security
    handler.go       # Report, ListReports, GetReport, Dashboard, Alerts, AckAlert
    store.go         # PostgreSQL CRUD
    model.go         # Report, Alert, DashboardSummary structs
    aggregator.go    # 跨節點匯總邏輯 (加權平均 compliance score)
```

### go-ot-security Agent 端改動

```
cmd/server/main.go   # 新增 flag: --coordinator-url, --node-id
internal/
  agent/
    reporter.go      # 掃描完成後 POST /security/report 到 coordinator
                     # 監控 cycle 完成後也自動上報
```

### Phase 3 完成標準
- [ ] ot-security 掃描結果 POST 到 dashai-go
- [ ] `/security/dashboard` 匯總多節點資料
- [ ] `/security/alerts` 跨節點告警列表
- [ ] 告警確認 (ack) 功能正常
- [ ] 歷史報告查詢正常

---

## Phase 4: Dashboard UI (選做)

在 dashai-go 嵌入一個輕量 React dashboard，顯示：
- Edge 節點地圖 (線上/離線/降級)
- 即時事件流 (WebSocket)
- 資安合規總覽 (4 框架分數)
- 設備拓撲 (依節點分組)

技術: React + shadcn/ui + Tailwind，Go embed 嵌入。
跟 go-ot-security dashboard 同樣的 dark theme 風格。

---

## 架構總覽

```
                          ┌─────────────────────┐
                          │   dashai-go (Cloud)  │
                          │   Render Starter     │
                          │                      │
                          │  /edge    ← 註冊心跳 │
                          │  /events  ← 事件匯流 │
                          │  /security← 資安匯總 │
                          │  /demo    ← 既有     │
                          │                      │
                          │  PostgreSQL (Neon)    │
                          │  WebSocket (即時推送)  │
                          └──────┬───────┬───────┘
                                 │       │
                    ┌────────────┘       └────────────┐
                    │                                  │
         ┌──────────┴──────────┐           ┌──────────┴──────────┐
         │  Factory A           │           │  Factory B           │
         │                      │           │                      │
         │  go-edge-gateway     │           │  go-edge-gateway     │
         │  ├─ Modbus plugin    │           │  ├─ SECS/GEM plugin  │
         │  ├─ MQTT plugin      │           │  └─ Modbus plugin    │
         │  └─ SECS/GEM plugin  │           │                      │
         │       ↕               │           │       ↕               │
         │  PLC  HMI  Equipment │           │  PLC  Equipment      │
         │                      │           │                      │
         │  go-ot-security      │           │  go-ot-security      │
         │  (資安掃描 agent)     │           │  (資安掃描 agent)     │
         └──────────────────────┘           └──────────────────────┘
```

## 時程估計

| Phase | 內容 | dashai-go | Agent 端 |
|-------|------|-----------|---------|
| 0 | Render 部署 | 恢復服務 | 新建 2 服務 |
| 1 | Edge 註冊心跳 | 5 檔案 | edge-gw 2 檔案 |
| 2 | 事件匯流 | 5 檔案 | edge-gw 改 2 檔案 |
| 3 | 資安匯總 | 5 檔案 | ot-sec 2 檔案 |
| 4 | Dashboard UI | React embed | 無 |

## 依賴關係

```
Phase 0 (部署) → Phase 1 (註冊) → Phase 2 (事件) → Phase 3 (資安)
                                                         ↓
                                                    Phase 4 (UI，選做)
```

## 合規提醒

- README/文案用 "implements" / "follows"，不用 "certified" / "compliant"
- AI-assisted development 標註
- 無 "production ready" — 用 "designed for factory deployment"
- IEC 62443 引用合規（標準編號可引用，不複製原文）

## 啟動詞

```
啟動分散式工廠平台 Phase 0: 部署 dashai-go + go-edge-gateway + go-factory-io 到 Render。
OpenSpec: ~/github/dashai-go/openspec/changes/distributed-factory-platform.md
```

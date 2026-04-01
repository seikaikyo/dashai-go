---
title: Scanner Frontend (Embedded React)
type: feature
status: completed
created: 2026-04-01
---

# Scanner Frontend (Embedded React)

## Background

Modbus scanner 目前只有 REST API，現場工程師要用 curl 操作不直覺。
需要一個嵌入式 Web UI，讓工程師在瀏覽器上操作掃描、查看點位表。

視覺風格參考 go-factory-io 的 SECSGEM Studio：暗色技術風 dashboard。

## Architecture

```
dashai-go/
├── web/
│   ├── scanner-ui/              # React 開發目錄
│   │   ├── src/
│   │   │   ├── App.tsx
│   │   │   ├── main.tsx
│   │   │   ├── components/
│   │   │   │   ├── ScanForm.tsx       # 掃描參數表單
│   │   │   │   ├── JobList.tsx        # 任務列表
│   │   │   │   ├── RegisterTable.tsx  # 點位表結果
│   │   │   │   ├── StatusBar.tsx      # 連線狀態
│   │   │   │   └── ExportButton.tsx   # CSV/JSON 匯出
│   │   │   └── lib/
│   │   │       └── api.ts            # API client
│   │   ├── package.json
│   │   ├── vite.config.ts
│   │   └── tailwind.config.ts
│   └── dist/                    # build 產出 (embedded)
├── internal/
│   └── scanner/
│       └── embed.go             # //go:embed web/dist/*
```

## Embed Pattern (same as go-factory-io)

```go
// internal/scanner/embed.go
//go:embed all:../../web/dist
var webFS embed.FS

// In router.go, mount at /scanner/
webSub, _ := fs.Sub(webFS, "web/dist")
r.Handle("/*", http.FileServer(http.FS(webSub)))
```

開發時: `cd web/scanner-ui && npm run dev` (Vite proxy to :8101)
部署時: `npm run build` → `go build` → 單一 binary

## UI Pages

### Main Dashboard (單頁四區塊)

```
┌──────────────────────────────────────────────┐
│  MODBUS SCANNER          [連線狀態] [Export]  │
├──────────────────────────────────────────────┤
│                                              │
│  ┌─ Scan Config ──────────────────────────┐  │
│  │ Host: [192.168.1.200] Port: [502]      │  │
│  │ Unit ID: [1]  Types: [x]H [x]I [ ]C   │  │
│  │ Range: [0] - [9999]  Samples: [5]      │  │
│  │                                        │  │
│  │ [Quick Scan]  [Full Scan]  [Read]      │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  ┌─ Jobs ────────────────────────────────┐   │
│  │ scan-1  192.168.1.200  completed  45s │   │
│  │ scan-2  192.168.1.201  running...     │   │
│  └───────────────────────────────────────┘   │
│                                              │
│  ┌─ Results ─────────────────────────────┐   │
│  │ Summary: 156 registers, 23 dynamic    │   │
│  │                                       │   │
│  │ Addr  Type     Value   Dynamic  Guess │   │
│  │ 0     uint16   253     yes      temp  │   │
│  │ 1     uint16   100     no       config│   │
│  │ 100   float32  25.35   yes      temp  │   │
│  │ ...                                   │   │
│  └───────────────────────────────────────┘   │
│                                              │
└──────────────────────────────────────────────┘
```

## Tech Stack

| Item | Choice |
|------|--------|
| Framework | React 19 + TypeScript |
| Styling | Tailwind CSS |
| Components | shadcn/ui (Table, Button, Input, Card, Badge, Select) |
| Build | Vite |
| Embed | Go `embed` package |

## Design Spec

參考 go-factory-io SECSGEM Studio 暗色風格：

| Token | Value |
|-------|-------|
| Background | `#0f1117` (dark) |
| Surface | `#1a1d27` (cards) |
| Border | `#2a2d37` |
| Text | `#e2e8f0` |
| Accent | `#3b82f6` (blue) |
| Success | `#22c55e` (green) |
| Warning | `#eab308` (yellow) |
| Error | `#ef4444` (red) |
| Font | `JetBrains Mono`, monospace |

## Features

### Scan Form
- Host/Port/Unit ID 輸入
- Scan types 多選 (holding/input/coil/discrete)
- Address range slider
- Samples 和 interval 設定
- Quick Scan / Full Scan / Read 按鈕

### Job List
- 所有掃描任務，5 秒自動 poll 更新
- 狀態 badge: running (blue pulse), completed (green), failed (red)
- 點擊展開結果

### Register Table
- 可排序/篩選的點位表
- 動態值高亮 (綠色閃爍)
- Float32 pair 標記
- Category badge (temperature/pressure/counter/config)
- 搜尋和篩選

### Export
- CSV 匯出 (給 Excel 用)
- JSON 匯出 (給程式用)
- 欄位: address, type, inferred_type, is_dynamic, value_range, guess

## Implementation Plan

### Phase 1: React 專案初始化
1. Vite + React + TypeScript
2. Tailwind + shadcn/ui
3. 暗色主題設定

### Phase 2: Components
1. ScanForm — 掃描參數表單
2. JobList — 任務列表 + auto poll
3. RegisterTable — 點位表 + sort/filter
4. StatusBar — 連線狀態
5. ExportButton — CSV/JSON 匯出

### Phase 3: API 整合
1. API client (fetch wrapper)
2. Vite dev proxy → localhost:8101

### Phase 4: Go Embed
1. embed.go 嵌入 dist/
2. router.go serve static files
3. 驗證: go build → 開 browser → 完整流程

## Checklist

- [x] Phase 1: React project setup
- [x] Phase 2: UI components
- [x] Phase 3: API integration
- [x] Phase 4: Go embed + verification

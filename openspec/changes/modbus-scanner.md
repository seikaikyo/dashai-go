---
title: Modbus Register Scanner Module
type: feature
status: completed
created: 2026-04-01
---

# Modbus Register Scanner Module

## Background

台灣工廠常見痛點：PLC 廠商不釋出點位表，或經過層層外包後文件遺失。
現場工程師面對一台 PLC，不知道哪些 register 有資料、代表什麼意義。

需要一個自動掃描工具，掃描 Modbus TCP 設備的所有 register，
產生點位表草稿，大幅縮短現場調試時間。

## Architecture

掛在 dashai-go 的 `/scanner` 路由下，提供 REST API：

```
dashai-go/
├── internal/
│   └── scanner/
│       ├── router.go          # HTTP API endpoints
│       ├── scanner.go         # 核心掃描邏輯
│       ├── analyzer.go        # 資料型態分析 + 動態值偵測
│       └── scanner_test.go    # 測試
```

## Scan Flow

```
使用者發 POST /scanner/api/scan
  → 連線目標 PLC (host:port)
  → Phase 1: Unit ID 探測 (1-247 可選)
  → Phase 2: Register 掃描 (holding / input / coil / discrete)
  → Phase 3: 連續採樣 N 次 (偵測動態值)
  → Phase 4: 資料型態推測 (int16 / uint16 / float32 / bool)
  → 回傳點位表 JSON
```

## API Endpoints

| Path | Method | Description |
|------|--------|-------------|
| `/scanner/api/scan` | POST | 啟動完整掃描 |
| `/scanner/api/scan/quick` | POST | 快速掃描 (只掃 holding register) |
| `/scanner/api/read` | POST | 讀取指定 register 範圍 |
| `/scanner/api/jobs` | GET | 查看進行中的掃描任務 |
| `/scanner/api/jobs/{id}` | GET | 查看特定任務結果 |

### POST /scanner/api/scan

```json
{
  "host": "192.168.1.200",
  "port": 502,
  "unit_id": 1,
  "scan_types": ["holding", "input", "coil"],
  "address_range": {"start": 0, "end": 9999},
  "batch_size": 125,
  "samples": 5,
  "sample_interval_ms": 1000,
  "timeout_ms": 500
}
```

### Response

```json
{
  "success": true,
  "data": {
    "job_id": "scan-abc123",
    "device": "192.168.1.200:502",
    "unit_id": 1,
    "scan_duration_ms": 45000,
    "summary": {
      "total_scanned": 10000,
      "responsive": 156,
      "dynamic": 23,
      "static": 133
    },
    "registers": [
      {
        "address": 0,
        "type": "holding",
        "raw_values": [253, 251, 255, 250, 252],
        "inferred_type": "int16",
        "is_dynamic": true,
        "value_range": {"min": 250, "max": 255},
        "guess": {
          "category": "temperature",
          "reason": "range 25.0-25.5 after /10 scaling, dynamic"
        }
      },
      {
        "address": 10,
        "type": "holding",
        "raw_values": [1, 1, 1, 1, 1],
        "inferred_type": "uint16",
        "is_dynamic": false,
        "value_range": {"min": 1, "max": 1},
        "guess": {
          "category": "config/status",
          "reason": "static value, small integer"
        }
      },
      {
        "address": 100,
        "type": "holding",
        "raw_values": [16968, 17000, 16950, 16990, 16970],
        "inferred_type": "float32_hi",
        "is_dynamic": true,
        "float32_value": 25.35,
        "paired_address": 101,
        "guess": {
          "category": "temperature",
          "reason": "float32 pair, value 25.35, dynamic"
        }
      }
    ]
  }
}
```

## Core Logic

### Phase 1: Unit ID Discovery

```
for uid := 1..247:
    send ReadHoldingRegisters(addr=0, count=1)
    if response → uid is alive
    if timeout → skip
```

### Phase 2: Register Scan

```
for addr := start..end step batch_size:
    send ReadHoldingRegisters(addr, min(batch_size, end-addr))
    record responsive addresses + raw values
    sleep(10ms) # 避免打爆 PLC
```

Modbus function codes:
- FC03: Read Holding Registers
- FC04: Read Input Registers
- FC01: Read Coils
- FC02: Read Discrete Inputs

### Phase 3: Multi-Sample

```
for sample := 1..N:
    re-read all responsive addresses
    record values
    sleep(sample_interval)

mark dynamic: any address where values changed across samples
```

### Phase 4: Type Inference

| Pattern | Inferred Type |
|---------|---------------|
| 值 0 或 1，coil | bool |
| 值 0-65535，單 register | uint16 |
| 值含負數 (-32768~32767) | int16 |
| 相鄰兩 register 組成 IEEE 754 float 在合理範圍 | float32 |
| 值固定不變 | config/parameter |
| 值持續變化 | sensor/measurement |

### Guess Categories

依值範圍和行為推測可能用途：

| 值範圍 (scaled) | Dynamic | Guess |
|-----------------|---------|-------|
| -40 ~ 200 | yes | temperature |
| 0 ~ 100 | yes | percentage (humidity/level) |
| 0 ~ 1000 | yes | pressure |
| 0 ~ 30000 | yes | rpm/speed |
| 0 或 1 | yes | on-off status |
| 小整數，固定 | no | config/mode |
| 持續遞增 | yes | counter/totalizer |

## Safety

- **唯讀操作**：只用 FC01-04 (Read)，絕對不寫入 (FC05/06/15/16)
- **速率控制**：batch 之間 sleep，避免打爆 PLC CPU
- **Timeout**：單次讀取 500ms timeout，不卡住
- **最大範圍**：預設掃 0-9999，可自訂

## Impact

| Item | Detail |
|------|--------|
| New files | 4 Go files in `internal/scanner/` |
| Dependencies | `github.com/goburrow/modbus` (already in go-edge-gateway) |
| Mount point | `/scanner` in main.go |

## Implementation Plan

### Phase 1: Core Scanner
1. `internal/scanner/scanner.go` — Modbus TCP 連線 + register 掃描
2. Batch read + timeout + rate limiting

### Phase 2: Analyzer
1. `internal/scanner/analyzer.go` — 多次採樣 + 動態偵測 + 型態推測
2. Float32 pair detection
3. Category guessing

### Phase 3: HTTP API
1. `internal/scanner/router.go` — REST API endpoints
2. Async job management (掃描可能花幾分鐘)
3. Mount to main.go

### Phase 4: Test
1. `internal/scanner/scanner_test.go` — analyzer 邏輯測試
2. 本機 curl 測試 (需要 Modbus simulator)

## Test Plan

| Phase | Test |
|-------|------|
| Analyzer | Unit test: type inference, dynamic detection, float32 pairing |
| API | curl POST /scanner/api/scan → 確認 JSON 格式正確 |
| Integration | diagslave simulator on localhost:502 |

## Checklist

- [x] Phase 1: Core Scanner (Modbus read + batch)
- [x] Phase 2: Analyzer (type inference + guessing)
- [x] Phase 3: HTTP API + job management
- [x] Phase 4: Tests (13 tests)
- [x] Mount to main.go

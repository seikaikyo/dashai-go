---
title: shukuyo-next 前端直接對接 Go API
type: feature
status: in-progress
created: 2026-04-03
---

# shukuyo-next 前端直接對接 Go API

## 變更內容

shukuyo-next (React) 目前打 Python API，改為直接打 dashai-go。
Go 需要補缺的端點，前端 hooks 改路徑和 response type。

## Go 缺少的端點（依優先序）

### P1: 靜態資料（知識庫）
- `GET /shukuyo/engine/mansions` — 27 宿列表
- `GET /shukuyo/engine/relations` — 6 組關係列表
- `GET /shukuyo/engine/metadata` — 歷史、來源、經典

### P2: 運勢擴展
- `POST /shukuyo/fortune/lucky-dates` — 吉日查詢
- `GET /shukuyo/fortune/calendar/{year}/{month}` — 月曆整合
- `GET /shukuyo/fortune/pair-lucky-days` — 配對吉日

### P3: 相性擴展
- `POST /shukuyo/engine/compatibility-batch` — 批次相性
- `GET /shukuyo/engine/compatibility-finder/{date}` — 相性 finder（27 宿全排列）

### P4: 創業模組
- `GET /shukuyo/fortune/startup/industry` — 行業推薦
- `GET /shukuyo/fortune/startup/lucky-calendar` — 創業吉日月曆

### P5: 外部整合（後做）
- `POST /shukuyo/career/104/company-jobs` — 104 職缺爬取
- `POST /shukuyo/career/company-search` — GCIS/全球公司搜尋

## 前端改動

改 10 個 hook 檔案的 API 路徑 + response type mapping：
- `src/config/api.ts` — base URL 指向 Go
- `src/hooks/use-fortune.ts`
- `src/hooks/use-company.ts`
- `src/hooks/use-compatibility.ts`
- `src/hooks/use-knowledge.ts`
- `src/hooks/use-calendar.ts`
- `src/hooks/use-lucky-days.ts`
- `src/hooks/use-pair-lucky-days.ts`
- `src/hooks/use-startup.ts`
- `src/hooks/use-mansion.ts`

## 影響範圍

### dashai-go 新增
- `internal/shukuyo/engine/handler.go` — 加 mansions/relations/metadata/batch/finder
- `internal/shukuyo/fortune/lucky.go` — 吉日計算
- `internal/shukuyo/fortune/calendar.go` — 月曆整合
- `internal/shukuyo/fortune/handler.go` — 新端點

### shukuyo-next 修改
- `src/config/api.ts` — base URL
- `src/hooks/*.ts` — 10 個 hook
- `src/types/*.ts` — response type 調整

## 測試計畫

1. Go 新端點 unit test
2. 前端 `npm run build` 通過
3. 本機啟動 Go server + next dev，手動走主流程
